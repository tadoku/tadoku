package domain

import (
	"fmt"
	"math"
	"sort"

	"github.com/google/uuid"
)

type ScoreSource string

const (
	ScoreSourceAmount          ScoreSource = "amount"
	ScoreSourceDurationMinutes ScoreSource = "duration_minutes"
)

type ScoringRuleSetMode string

const (
	ScoringRuleSetModeOverride ScoringRuleSetMode = "override"
	ScoringRuleSetModeReplace  ScoringRuleSetMode = "replace"
)

type ScoringRule struct {
	ID           uuid.UUID
	Priority     int32
	Stackable    bool
	ActivityID   int32
	UnitKey      string
	LanguageCode string
	Tag          string
	ScoreSource  ScoreSource
	Rate         float32
}

type ScoringRuleSet struct {
	ID      uuid.UUID
	Version int32
	Mode    ScoringRuleSetMode
	Rules   []ScoringRule
}

type ScoringInput struct {
	ActivityID      int32
	UnitKey         string
	LanguageCode    string
	Tags            []string
	Amount          *float32
	DurationSeconds *int32
}

type AppliedScoringRule struct {
	RuleID uuid.UUID
	Rate   float32
}

type ScoringResult struct {
	Score            float32
	ScoreSource      ScoreSource
	Matched          bool
	AppliedRuleSetID *uuid.UUID
	AppliedRules     []AppliedScoringRule
}

// EvaluateScoringRuleSet applies the first matching base rule in ascending
// priority order, then multiplies the rates of every matching stackable rule.
func EvaluateScoringRuleSet(input ScoringInput, ruleSet ScoringRuleSet) (ScoringResult, error) {
	source, scoreableValue, err := scoringSourceAndValue(input)
	if err != nil {
		return ScoringResult{}, err
	}
	if err := validateScoringRules(ruleSet.Rules); err != nil {
		return ScoringResult{}, err
	}

	result := ScoringResult{
		Score:       scoreableValue,
		ScoreSource: source,
	}
	rules := append([]ScoringRule(nil), ruleSet.Rules...)
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	tags := make(map[string]struct{}, len(input.Tags))
	for _, tag := range input.Tags {
		tags[tag] = struct{}{}
	}

	var baseRule *ScoringRule
	modifierRules := make([]ScoringRule, 0)
	for i := range rules {
		rule := &rules[i]
		if !scoringRuleMatches(*rule, input, source, tags) {
			continue
		}

		if rule.Stackable {
			modifierRules = append(modifierRules, *rule)
		} else if baseRule == nil {
			baseRule = rule
		}
	}

	if baseRule == nil {
		result.Score = 0
		return result, nil
	}

	result.Score *= baseRule.Rate
	result.AppliedRules = append(result.AppliedRules, AppliedScoringRule{
		RuleID: baseRule.ID,
		Rate:   baseRule.Rate,
	})
	for _, rule := range modifierRules {
		result.Score *= rule.Rate
		result.AppliedRules = append(result.AppliedRules, AppliedScoringRule{
			RuleID: rule.ID,
			Rate:   rule.Rate,
		})
	}

	ruleSetID := ruleSet.ID
	result.Matched = true
	result.AppliedRuleSetID = &ruleSetID
	return result, nil
}

// EvaluateContestScore evaluates a contest rule set according to its mode.
// Override sets fall back to their pinned platform set; replace sets do not.
func EvaluateContestScore(
	input ScoringInput,
	contestRuleSet ScoringRuleSet,
	fallbackRuleSet *ScoringRuleSet,
) (ScoringResult, error) {
	result, err := EvaluateScoringRuleSet(input, contestRuleSet)
	if err != nil {
		return ScoringResult{}, err
	}
	if result.Matched {
		return result, nil
	}

	switch contestRuleSet.Mode {
	case ScoringRuleSetModeReplace:
		return result, nil
	case ScoringRuleSetModeOverride:
		if fallbackRuleSet == nil {
			return ScoringResult{}, fmt.Errorf("override scoring rule set requires a fallback: %w", ErrInvalidScoringRuleSet)
		}
		return EvaluateScoringRuleSet(input, *fallbackRuleSet)
	default:
		return ScoringResult{}, fmt.Errorf("unknown contest scoring rule set mode %q: %w", contestRuleSet.Mode, ErrInvalidScoringRuleSet)
	}
}

func scoringSourceAndValue(input ScoringInput) (ScoreSource, float32, error) {
	if !IsValidActivityID(input.ActivityID) {
		return "", 0, fmt.Errorf("activity %d is not valid: %w", input.ActivityID, ErrInvalidLog)
	}
	if input.Amount != nil && (!isFiniteFloat32(*input.Amount) || *input.Amount <= 0) {
		return "", 0, fmt.Errorf("amount must be positive and finite: %w", ErrInvalidLog)
	}
	if input.DurationSeconds != nil && *input.DurationSeconds <= 0 {
		return "", 0, fmt.Errorf("duration_seconds must be positive: %w", ErrInvalidLog)
	}
	if input.Amount != nil {
		return ScoreSourceAmount, *input.Amount, nil
	}
	if input.DurationSeconds != nil {
		return ScoreSourceDurationMinutes, float32(*input.DurationSeconds) / 60, nil
	}
	return "", 0, fmt.Errorf("amount or duration_seconds is required: %w", ErrInvalidLog)
}

func validateScoringRules(rules []ScoringRule) error {
	priorities := make(map[int32]struct{}, len(rules))
	for _, rule := range rules {
		if _, exists := priorities[rule.Priority]; exists {
			return fmt.Errorf("duplicate scoring rule priority %d: %w", rule.Priority, ErrInvalidScoringRuleSet)
		}
		priorities[rule.Priority] = struct{}{}

		if !IsValidActivityID(rule.ActivityID) {
			return fmt.Errorf("scoring rule activity %d is not valid: %w", rule.ActivityID, ErrInvalidScoringRuleSet)
		}
		if rule.ScoreSource != ScoreSourceAmount && rule.ScoreSource != ScoreSourceDurationMinutes {
			return fmt.Errorf("scoring rule score source %q is not valid: %w", rule.ScoreSource, ErrInvalidScoringRuleSet)
		}
		if !isFiniteFloat32(rule.Rate) || rule.Rate < 0 {
			return fmt.Errorf("scoring rule rate must be non-negative and finite: %w", ErrInvalidScoringRuleSet)
		}
	}
	return nil
}

func scoringRuleMatches(
	rule ScoringRule,
	input ScoringInput,
	source ScoreSource,
	tags map[string]struct{},
) bool {
	if rule.ScoreSource != source || rule.ActivityID != input.ActivityID {
		return false
	}
	if rule.UnitKey != "" && rule.UnitKey != input.UnitKey {
		return false
	}
	if rule.LanguageCode != "" && rule.LanguageCode != input.LanguageCode {
		return false
	}
	if rule.Tag != "" {
		if _, ok := tags[rule.Tag]; !ok {
			return false
		}
	}
	return true
}

func isFiniteFloat32(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}
