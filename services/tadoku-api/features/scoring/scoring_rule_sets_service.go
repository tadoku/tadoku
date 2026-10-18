package scoring

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	domainlanguages "github.com/tadoku/tadoku/services/tadoku-api/domain/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

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

func (s *Service) validateContestDraftConfiguration(ctx context.Context, parameters *ContestDraftParameters) error {
	var fallbackID uuid.UUID
	switch configuration := parameters.Configuration.(type) {
	case Replace:
		return nil
	case Override:
		fallbackID = configuration.FallbackRuleSetID
	default:
		return errx.NewInternalError("invalid contest draft configuration")
	}

	fallback, err := s.repository.FindRuleSetByID(ctx, fallbackID)
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
	if err := s.validateContestDraftConfiguration(ctx, &parameters); err != nil {
		return nil, err
	}

	draft := RuleSet{
		Scope:     "contest",
		ContestID: &parameters.ContestID,
		Rules:     parameters.Rules,
		CreatedAt: parameters.CreatedAt,
	}
	switch configuration := parameters.Configuration.(type) {
	case Replace:
		draft.Mode = string(modeReplace)
	case Override:
		draft.Mode = string(modeOverride)
		draft.FallbackRuleSetID = &configuration.FallbackRuleSetID
	default:
		return nil, errx.NewInternalError("invalid contest draft configuration")
	}

	return s.createDraft(ctx, draft)
}

func (s *Service) nextDraftVersion(ctx context.Context, contestID *uuid.UUID) (int32, error) {
	if contestID == nil {
		return s.repository.NextPlatformDraftVersion(ctx)
	}
	return s.repository.NextContestDraftVersion(ctx, *contestID)
}

func (s *Service) createDraft(ctx context.Context, draft RuleSet) (*RuleSet, error) {
	version, err := s.nextDraftVersion(ctx, draft.ContestID)
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

func (s *Service) FindRuleSet(ctx context.Context, id uuid.UUID) (*RuleSet, error) {
	return s.repository.FindRuleSetByID(ctx, id)
}

func (s *Service) PublishRuleSet(ctx context.Context, ruleSet RuleSet, publishedAt time.Time) (*RuleSet, error) {
	if ruleSet.Status != "draft" {
		return nil, errx.NewConflictError("only draft scoring rule sets can be published")
	}

	published, err := s.repository.PublishRuleSet(ctx, ruleSet.ID, publishedAt)
	if err != nil {
		return nil, err
	}
	published.Rules, err = s.repository.ListRules(ctx, published.ID)
	return published, err
}

func (s *Service) ActivateRuleSet(ctx context.Context, ruleSet RuleSet, updatedAt time.Time) error {
	if ruleSet.Status != "published" {
		return errx.NewConflictError("only published scoring rule sets can be activated")
	}

	switch ruleSet.Scope {
	case "platform":
		return s.repository.ActivatePlatformRuleSet(ctx, ruleSet.ID)
	case "contest":
		if ruleSet.ContestID == nil {
			return errx.NewInvalidInputError("contest scoring rule set requires contest_id")
		}
		return s.repository.ActivateContestRuleSet(ctx, *ruleSet.ContestID, ruleSet.ID, updatedAt)
	default:
		return errx.NewInvalidInputError("scoring rule set scope is invalid")
	}
}
