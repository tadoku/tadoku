package scoring

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	domainlanguages "github.com/tadoku/tadoku/services/tadoku-api/domain/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Service struct{ repository *ScoringRepository }

func NewService(repository *ScoringRepository) *Service { return &Service{repository: repository} }

func (s *Service) ListPlatformRuleSets(ctx context.Context, includeDrafts bool) ([]RuleSet, error) {
	sets, err := s.repository.ListPlatformRuleSets(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.hydrateRuleSets(ctx, sets); err != nil {
		return nil, err
	}
	active, err := s.repository.FindActivePlatformRuleSet(ctx)
	if err != nil {
		return nil, err
	}
	for i := range sets {
		sets[i].Active = sets[i].ID == active.ID
	}
	if includeDrafts {
		return sets, nil
	}
	published := make([]RuleSet, 0, len(sets))
	for _, set := range sets {
		if set.Status == "published" {
			published = append(published, set)
		}
	}
	return published, nil
}

func (s *Service) ListContestRuleSets(ctx context.Context, contestID uuid.UUID) ([]RuleSet, error) {
	sets, err := s.repository.ListContestRuleSets(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if err := s.hydrateRuleSets(ctx, sets); err != nil {
		return nil, err
	}
	activeID, err := s.repository.FindContestActiveRuleSetID(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if activeID != nil {
		for i := range sets {
			sets[i].Active = sets[i].ID == *activeID
		}
	}
	return sets, nil
}

func (s *Service) Preview(ctx context.Context, parameters PreviewParameters) (*Preview, error) {
	unitKey, err := s.resolveUnit(ctx, parameters)
	if err != nil {
		return nil, err
	}
	tags, err := NormalizeTags(parameters.Tags)
	if err != nil {
		return nil, err
	}

	input := scoringInput{
		activityID:      parameters.ActivityID,
		unitKey:         unitKey,
		languageCode:    parameters.LanguageCode,
		tags:            tags,
		amount:          parameters.Amount,
		durationSeconds: parameters.DurationSeconds,
	}

	platformSet, err := s.repository.FindActivePlatformRuleSet(ctx)
	if err != nil {
		return nil, fmt.Errorf("preview platform score: %w", err)
	}
	platformSet.Rules, err = s.repository.ListRules(ctx, platformSet.ID)
	if err != nil {
		return nil, err
	}
	platform, _, err := evaluate(input, *platformSet)
	if err != nil {
		return nil, fmt.Errorf("preview platform score: %w", err)
	}

	result := &Preview{Platform: platform, Contests: make([]ContestEstimate, 0, len(parameters.Contests))}
	for _, contest := range parameters.Contests {
		set, fallback, err := s.findContestRuleSets(ctx, contest.ContestID)
		if err != nil {
			return nil, fmt.Errorf("preview contest score: %w", err)
		}
		estimate := platform
		if set != nil {
			var matched bool
			estimate, matched, err = evaluate(input, *set)
			if err != nil {
				return nil, fmt.Errorf("preview contest score: %w", err)
			}
			if !matched && set.Mode == "override" {
				if fallback == nil {
					return nil, errx.NewInternalError("override scoring rule set requires a fallback")
				}
				estimate, _, err = evaluate(input, *fallback)
			}
			if !matched && set.Mode != "override" && set.Mode != "replace" {
				return nil, errx.NewInternalError("unknown contest scoring mode")
			}
			if err != nil {
				return nil, fmt.Errorf("preview contest score: %w", err)
			}
		}
		result.Contests = append(result.Contests, ContestEstimate{
			RegistrationID: contest.RegistrationID,
			ContestID:      contest.ContestID,
			Estimate:       estimate,
		})
	}
	return result, nil
}

func (s *Service) resolveUnit(ctx context.Context, parameters PreviewParameters) (string, error) {
	if !validActivity(parameters.ActivityID) {
		return "", errx.NewInvalidInputError("activity_id is not valid")
	}
	if parameters.UnitKey != nil && unitActivities[*parameters.UnitKey] != parameters.ActivityID {
		return "", errx.NewInvalidInputError("unit_key is not valid for activity_id")
	}
	hasUnit := parameters.UnitID != nil || parameters.UnitKey != nil
	if (parameters.Amount != nil) != hasUnit {
		return "", errx.NewInvalidInputError("amount and a unit identifier must be supplied together")
	}
	if parameters.DurationSeconds != nil && *parameters.DurationSeconds <= 0 {
		return "", errx.NewInvalidInputError("duration_seconds must be positive")
	}
	if parameters.Amount != nil && (!finite(*parameters.Amount) || *parameters.Amount <= 0) {
		return "", errx.NewInvalidInputError("amount must be positive and finite")
	}
	if parameters.Amount == nil && parameters.DurationSeconds == nil {
		return "", errx.NewInvalidInputError("amount/unit or duration_seconds is required")
	}
	if parameters.Amount == nil {
		return "", nil
	}

	var resolved string
	if parameters.UnitID != nil {
		var err error
		resolved, err = s.repository.FindUnitKeyByID(ctx, *parameters.UnitID, parameters.ActivityID, parameters.LanguageCode)
		if err != nil {
			return "", err
		}
	}
	if parameters.UnitID == nil && parameters.UnitKey != nil {
		var err error
		resolved, err = s.repository.FindUnitKeyByKey(ctx, *parameters.UnitKey, parameters.ActivityID, parameters.LanguageCode)
		if err != nil {
			return "", err
		}
	}
	if parameters.UnitKey != nil && resolved != *parameters.UnitKey {
		return "", errx.NewInvalidInputError("unit_id and unit_key identify different units")
	}
	if unitActivities[resolved] != parameters.ActivityID {
		return "", errx.NewInvalidInputError("resolved unit_key is not valid for activity_id")
	}
	return resolved, nil
}

func (s *Service) hydrateRuleSets(ctx context.Context, sets []RuleSet) error {
	for i := range sets {
		rules, err := s.repository.ListRules(ctx, sets[i].ID)
		if err != nil {
			return err
		}
		sets[i].Rules = rules
	}
	return nil
}

func (s *Service) findContestRuleSets(ctx context.Context, contestID uuid.UUID) (*RuleSet, *RuleSet, error) {
	id, err := s.repository.FindContestActiveRuleSetID(ctx, contestID)
	if err != nil || id == nil {
		return nil, nil, err
	}
	set, err := s.repository.FindRuleSetByID(ctx, *id)
	if err != nil {
		return nil, nil, err
	}
	set.Rules, err = s.repository.ListRules(ctx, set.ID)
	if err != nil || set.FallbackRuleSetID == nil {
		return set, nil, err
	}
	fallback, err := s.repository.FindRuleSetByID(ctx, *set.FallbackRuleSetID)
	if err != nil {
		return nil, nil, err
	}
	fallback.Rules, err = s.repository.ListRules(ctx, fallback.ID)
	return set, fallback, err
}

func (s *Service) ValidateContestDraftConfiguration(ctx context.Context, parameters *ContestDraftParameters) error {
	fallbackID := parameters.Configuration.FallbackRuleSetID()
	if fallbackID == nil {
		return nil
	}

	fallback, err := s.repository.FindRuleSetByID(ctx, *fallbackID)
	if err != nil {
		return err
	}
	if fallback.Scope != "platform" || fallback.Status != "published" {
		return errx.NewInvalidInputError("fallback must be a published platform rule set")
	}

	return nil
}

func (s *Service) NormalizeDraftRules(parameters *DraftRules, languages []domainlanguages.Language) error {
	return parameters.normalize(languages)
}

func (s *Service) CreatePlatformDraft(ctx context.Context, parameters PlatformDraftParameters) (*RuleSet, error) {
	return s.createDraft(ctx, RuleSet{
		Scope:     "platform",
		Rules:     parameters.Rules,
		CreatedAt: parameters.CreatedAt,
	})
}

func (s *Service) CreateContestDraft(ctx context.Context, parameters ContestDraftParameters) (*RuleSet, error) {
	return s.createDraft(ctx, RuleSet{
		Scope:             "contest",
		ContestID:         &parameters.ContestID,
		Mode:              string(parameters.Configuration.Mode()),
		FallbackRuleSetID: parameters.Configuration.FallbackRuleSetID(),
		Rules:             parameters.Rules,
		CreatedAt:         parameters.CreatedAt,
	})
}

func (s *Service) createDraft(ctx context.Context, draft RuleSet) (*RuleSet, error) {
	version, err := s.repository.NextDraftVersion(ctx, draft.ContestID)
	if err != nil {
		return nil, fmt.Errorf("allocate scoring rule set version: %w", err)
	}

	draft.ID = uuid.New()
	draft.Version = version
	draft.Status = "draft"
	rules := draft.Rules
	draft.Rules = []Rule{}

	created, err := s.repository.CreateDraft(ctx, draft)
	if err != nil {
		return nil, err
	}
	created.Rules = make([]Rule, len(rules))
	for i, rule := range rules {
		rule.ID = uuid.New()
		if err := s.repository.CreateRule(ctx, created.ID, rule); err != nil {
			return nil, err
		}
		created.Rules[i] = rule
	}

	return created, nil
}
