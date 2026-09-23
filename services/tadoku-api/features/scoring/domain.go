package scoring

import (
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var (
	ErrRuleSetNotFound = errors.New("scoring rule set not found")
	ErrContestNotFound = errx.NewNotFoundError("contest not found")
)

type Source string

const (
	SourceAmount          Source = "amount"
	SourceDurationMinutes Source = "duration_minutes"
)

type Mode string

const (
	ModeReplace  Mode = "replace"
	ModeOverride Mode = "override"
)

type Rule struct {
	ID           uuid.UUID
	Priority     int32
	Stackable    bool
	ActivityID   int32
	UnitKey      string
	LanguageCode string
	Tag          string
	Source       Source
	Rate         float32
}

type RuleSet struct {
	ID                uuid.UUID
	Scope             string
	ContestID         *uuid.UUID
	Version           int32
	Status            string
	Active            bool
	Mode              string
	FallbackRuleSetID *uuid.UUID
	Rules             []Rule
	CreatedAt         time.Time
	PublishedAt       *time.Time
}

type DraftParameters struct {
	ContestID         *uuid.UUID
	Mode              Mode
	FallbackRuleSetID *uuid.UUID
	Rules             []Rule
	LanguageCodes     map[string]struct{}
	CreatedAt         time.Time
}

func ParseMode(value string) (Mode, error) {
	mode := Mode(value)
	if mode != ModeReplace && mode != ModeOverride {
		return "", errx.NewInvalidInputError("contest scoring mode is required")
	}

	return mode, nil
}

func (p *DraftParameters) validateConfiguration(scope string) error {
	if scope == "contest" {
		switch p.Mode {
		case ModeReplace:
			if p.FallbackRuleSetID != nil {
				return errx.NewInvalidInputError("replace rule sets cannot have a fallback")
			}
		case ModeOverride:
			if p.FallbackRuleSetID == nil {
				return errx.NewInvalidInputError("override rule sets require a fallback")
			}
		}
	}
	return nil
}

func (p *DraftParameters) validateRules() error {
	priorities := make(map[int32]struct{}, len(p.Rules))
	for i := range p.Rules {
		rule := &p.Rules[i]
		rule.Tag = strings.ToLower(strings.TrimSpace(rule.Tag))
		rule.LanguageCode = strings.ToLower(strings.TrimSpace(rule.LanguageCode))
		if rule.Priority < 0 {
			return errx.NewInvalidInputError("rule priority must be non-negative")
		}
		if _, duplicate := priorities[rule.Priority]; duplicate {
			return errx.NewInvalidInputError("rule priorities must be unique")
		}
		priorities[rule.Priority] = struct{}{}
		if !validActivity(rule.ActivityID) {
			return errx.NewInvalidInputError("rule activity_id is not valid")
		}
		if rule.UnitKey != "" && unitActivities[rule.UnitKey] != rule.ActivityID {
			return errx.NewInvalidInputError("rule unit_key is not valid for activity_id")
		}
		if len(rule.LanguageCode) > 10 {
			return errx.NewInvalidInputError("rule language_code is too long")
		}
		if rule.LanguageCode != "" {
			if _, exists := p.LanguageCodes[rule.LanguageCode]; !exists {
				return errx.NewInvalidInputError("rule language_code is not valid")
			}
		}
		if rule.Tag != "" {
			tags, err := NormalizeTags([]string{rule.Tag})
			if err != nil || len(tags) != 1 {
				return errx.NewInvalidInputError("rule tag is invalid")
			}
			rule.Tag = tags[0]
		}
		if rule.Source != SourceAmount && rule.Source != SourceDurationMinutes {
			return errx.NewInvalidInputError("rule score_source is not valid")
		}
		if !finite(rule.Rate) || rule.Rate < 0 {
			return errx.NewInvalidInputError("rule rate must be non-negative and finite")
		}
	}
	return nil
}

type PreviewParameters struct {
	UnitID          *uuid.UUID
	UnitKey         *string
	ActivityID      int32
	LanguageCode    string
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
	Contests        []PreviewContest
}

type PreviewContest struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
}

type Estimate struct {
	Score     float32
	Source    Source
	RuleSetID *uuid.UUID
	Rules     []AppliedRule
}

type AppliedRule struct {
	RuleID uuid.UUID
	Rate   float32
}

type ContestEstimate struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
	Estimate       Estimate
}

type Preview struct {
	Platform Estimate
	Contests []ContestEstimate
}

type scoringInput struct {
	activityID      int32
	unitKey         string
	languageCode    string
	tags            []string
	amount          *float32
	durationSeconds *int32
}

func ValidatePreview(parameters PreviewParameters) error {
	if parameters.ActivityID == 0 {
		return errx.NewInvalidInputError("activity_id is required")
	}
	if parameters.LanguageCode == "" {
		return errx.NewInvalidInputError("language_code is required")
	}

	_, err := NormalizeTags(parameters.Tags)
	if err != nil {
		return err
	}

	return nil
}

func evaluate(input scoringInput, ruleSet RuleSet) (Estimate, bool, error) {
	value, source, err := scoreableValue(input)
	if err != nil {
		return Estimate{}, false, err
	}

	rules := append([]Rule(nil), ruleSet.Rules...)
	sort.SliceStable(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })

	tags := make(map[string]struct{}, len(input.tags))
	for _, tag := range input.tags {
		tags[tag] = struct{}{}
	}

	var base *Rule
	modifiers := make([]Rule, 0)
	priorities := make(map[int32]struct{}, len(rules))
	for i := range rules {
		rule := &rules[i]
		if err := validateStoredRule(*rule, priorities); err != nil {
			return Estimate{}, false, err
		}
		priorities[rule.Priority] = struct{}{}

		if !ruleMatches(*rule, input, source, tags) {
			continue
		}
		if rule.Stackable {
			modifiers = append(modifiers, *rule)
		} else if base == nil {
			base = rule
		}
	}

	estimate := Estimate{
		Score:  0,
		Source: source,
		Rules:  []AppliedRule{},
	}
	if base == nil {
		return estimate, false, nil
	}
	estimate.Score = value * base.Rate
	estimate.RuleSetID = &ruleSet.ID
	estimate.Rules = append(estimate.Rules, AppliedRule{
		RuleID: base.ID,
		Rate:   base.Rate,
	})
	for _, rule := range modifiers {
		estimate.Score *= rule.Rate
		estimate.Rules = append(estimate.Rules, AppliedRule{
			RuleID: rule.ID,
			Rate:   rule.Rate,
		})
	}
	return estimate, true, nil
}

func validateStoredRule(rule Rule, priorities map[int32]struct{}) error {
	if _, duplicate := priorities[rule.Priority]; duplicate {
		return errx.NewInternalError("duplicate rule priority")
	}
	if !validActivity(rule.ActivityID) {
		return errx.NewInternalError("invalid rule activity")
	}
	if rule.Source != SourceAmount && rule.Source != SourceDurationMinutes {
		return errx.NewInternalError("invalid rule source")
	}
	if !finite(rule.Rate) || rule.Rate < 0 {
		return errx.NewInternalError("invalid rule rate")
	}
	return nil
}

func ruleMatches(rule Rule, input scoringInput, source Source, tags map[string]struct{}) bool {
	if rule.Source != source || rule.ActivityID != input.activityID {
		return false
	}
	if rule.UnitKey != "" && rule.UnitKey != input.unitKey {
		return false
	}
	if rule.LanguageCode != "" && rule.LanguageCode != input.languageCode {
		return false
	}
	if rule.Tag != "" {
		_, matches := tags[rule.Tag]
		return matches
	}
	return true
}

func scoreableValue(input scoringInput) (float32, Source, error) {
	if !validActivity(input.activityID) {
		return 0, "", errx.NewInvalidInputError("activity_id is not valid")
	}
	if input.languageCode == "" {
		return 0, "", errx.NewInvalidInputError("language_code is required")
	}
	if input.amount != nil {
		if !finite(*input.amount) || *input.amount <= 0 {
			return 0, "", errx.NewInvalidInputError("amount must be positive and finite")
		}
		return *input.amount, SourceAmount, nil
	}
	if input.durationSeconds != nil && *input.durationSeconds > 0 {
		return float32(*input.durationSeconds) / 60, SourceDurationMinutes, nil
	}
	return 0, "", errx.NewInvalidInputError("amount or duration_seconds is required")
}

func validActivity(id int32) bool {
	all := activities.All()
	return id > 0 && int(id) <= len(all) && all[id-1].ID == id
}

var unitActivities = map[string]int32{
	"reading_page":            1,
	"reading_two_column_page": 1,
	"reading_comic_page":      1,
	"reading_sentence":        1,
	"reading_character":       1,
	"listening_minute":        2,
	"listening_dense_minutes": 2,
	"writing_page":            3,
	"writing_sentence":        3,
	"writing_character":       3,
	"speaking_minute":         4,
	"speaking_dense_minutes":  4,
	"study_minute":            5,
}

func finite(value float32) bool { return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0) }

func NormalizeTags(tags []string) ([]string, error) {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if len(tag) > 50 {
			return nil, errx.NewInternalError("tag exceeds maximum length of 50 characters")
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	if len(result) > 10 {
		return nil, errx.NewInternalError("more than 10 tags remain after normalization")
	}
	return result, nil
}
