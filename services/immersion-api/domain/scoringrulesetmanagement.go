package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type ScoringRuleSetManagementRepository interface {
	FindContestByID(context.Context, *ContestFindRequest) (*ContestView, error)
	FindScoringRuleSetByID(context.Context, uuid.UUID) (*ScoringRuleSet, error)
	ListLanguages(context.Context) ([]Language, error)
	ListPlatformScoringRuleSets(context.Context) ([]ScoringRuleSet, error)
	ListContestScoringRuleSets(context.Context, uuid.UUID) ([]ScoringRuleSet, error)
	CreateScoringRuleSetDraft(context.Context, *ScoringRuleSetDraftCreateRequest) (*ScoringRuleSet, error)
	PublishScoringRuleSet(context.Context, uuid.UUID, time.Time) (*ScoringRuleSet, error)
	ActivatePlatformScoringRuleSet(context.Context, uuid.UUID) error
	ActivateContestScoringRuleSet(context.Context, uuid.UUID, uuid.UUID, time.Time) error
}

type ScoringRuleSetDraftCreateRequest struct {
	ContestID         *uuid.UUID
	Mode              ScoringRuleSetMode
	FallbackRuleSetID *uuid.UUID
	Rules             []ScoringRule

	scope ScoringRuleSetScope
}

func (r *ScoringRuleSetDraftCreateRequest) Scope() ScoringRuleSetScope { return r.scope }

type ScoringRuleSetManagement struct {
	repo  ScoringRuleSetManagementRepository
	clock commondomain.Clock
}

func NewScoringRuleSetManagement(
	repo ScoringRuleSetManagementRepository,
	clock commondomain.Clock,
) *ScoringRuleSetManagement {
	return &ScoringRuleSetManagement{repo: repo, clock: clock}
}

func (s *ScoringRuleSetManagement) ListPlatform(ctx context.Context) ([]ScoringRuleSet, error) {
	if err := requireAuthentication(ctx); err != nil {
		return nil, err
	}
	ruleSets, err := s.repo.ListPlatformScoringRuleSets(ctx)
	if err != nil {
		return nil, err
	}
	if isAdmin(ctx) {
		return ruleSets, nil
	}
	published := make([]ScoringRuleSet, 0, len(ruleSets))
	for _, ruleSet := range ruleSets {
		if ruleSet.Status == ScoringRuleSetStatusPublished {
			published = append(published, ruleSet)
		}
	}
	return published, nil
}

func (s *ScoringRuleSetManagement) ListContest(
	ctx context.Context,
	contestID uuid.UUID,
) ([]ScoringRuleSet, error) {
	if _, err := s.requireContestOwner(ctx, contestID); err != nil {
		return nil, err
	}
	return s.repo.ListContestScoringRuleSets(ctx, contestID)
}

func (s *ScoringRuleSetManagement) CreatePlatformDraft(
	ctx context.Context,
	req *ScoringRuleSetDraftCreateRequest,
) (*ScoringRuleSet, error) {
	if !isAdmin(ctx) {
		return nil, ErrForbidden
	}
	req.scope = ScoringRuleSetScopePlatform
	req.ContestID = nil
	req.Mode = ""
	req.FallbackRuleSetID = nil
	if err := s.validateDraft(ctx, req); err != nil {
		return nil, err
	}
	return s.repo.CreateScoringRuleSetDraft(ctx, req)
}

func (s *ScoringRuleSetManagement) CreateContestDraft(
	ctx context.Context,
	contestID uuid.UUID,
	req *ScoringRuleSetDraftCreateRequest,
) (*ScoringRuleSet, error) {
	if err := s.requireContestOwnerBeforeStart(ctx, contestID); err != nil {
		return nil, err
	}
	req.scope = ScoringRuleSetScopeContest
	req.ContestID = &contestID
	if err := s.validateDraft(ctx, req); err != nil {
		return nil, err
	}
	return s.repo.CreateScoringRuleSetDraft(ctx, req)
}

func (s *ScoringRuleSetManagement) Publish(
	ctx context.Context,
	ruleSetID uuid.UUID,
) (*ScoringRuleSet, error) {
	ruleSet, err := s.repo.FindScoringRuleSetByID(ctx, ruleSetID)
	if err != nil {
		return nil, err
	}
	if err := s.authorizeRuleSetChange(ctx, ruleSet); err != nil {
		return nil, err
	}
	if ruleSet.Status != ScoringRuleSetStatusDraft {
		return nil, fmt.Errorf("only draft rule sets can be published: %w", ErrConflict)
	}
	return s.repo.PublishScoringRuleSet(ctx, ruleSetID, s.clock.Now())
}

func (s *ScoringRuleSetManagement) Activate(
	ctx context.Context,
	ruleSetID uuid.UUID,
) error {
	ruleSet, err := s.repo.FindScoringRuleSetByID(ctx, ruleSetID)
	if err != nil {
		return err
	}
	if err := s.authorizeRuleSetChange(ctx, ruleSet); err != nil {
		return err
	}
	if ruleSet.Status != ScoringRuleSetStatusPublished {
		return fmt.Errorf("only published rule sets can be activated: %w", ErrConflict)
	}
	switch ruleSet.Scope {
	case ScoringRuleSetScopePlatform:
		return s.repo.ActivatePlatformScoringRuleSet(ctx, ruleSetID)
	case ScoringRuleSetScopeContest:
		if ruleSet.ContestID == nil {
			return ErrInvalidScoringRuleSet
		}
		return s.repo.ActivateContestScoringRuleSet(ctx, *ruleSet.ContestID, ruleSetID, s.clock.Now())
	default:
		return ErrInvalidScoringRuleSet
	}
}

func (s *ScoringRuleSetManagement) validateDraft(
	ctx context.Context,
	req *ScoringRuleSetDraftCreateRequest,
) error {
	if req.scope == ScoringRuleSetScopeContest {
		switch req.Mode {
		case ScoringRuleSetModeReplace:
			if req.FallbackRuleSetID != nil {
				return fmt.Errorf("replace rule sets cannot have a fallback: %w", ErrInvalidScoringRuleSet)
			}
		case ScoringRuleSetModeOverride:
			if req.FallbackRuleSetID == nil {
				return fmt.Errorf("override rule sets require a fallback: %w", ErrInvalidScoringRuleSet)
			}
			fallback, err := s.repo.FindScoringRuleSetByID(ctx, *req.FallbackRuleSetID)
			if err != nil {
				return err
			}
			if fallback.Scope != ScoringRuleSetScopePlatform || fallback.Status != ScoringRuleSetStatusPublished {
				return fmt.Errorf("fallback must be a published platform rule set: %w", ErrInvalidScoringRuleSet)
			}
		default:
			return fmt.Errorf("contest scoring mode is required: %w", ErrInvalidScoringRuleSet)
		}
	}
	languages, err := s.repo.ListLanguages(ctx)
	if err != nil {
		return fmt.Errorf("could not validate scoring rule languages: %w", err)
	}
	languageCodes := make(map[string]struct{}, len(languages))
	for _, language := range languages {
		languageCodes[language.Code] = struct{}{}
	}
	for i := range req.Rules {
		rule := &req.Rules[i]
		rule.Tag = strings.ToLower(strings.TrimSpace(rule.Tag))
		rule.LanguageCode = strings.ToLower(strings.TrimSpace(rule.LanguageCode))
		if rule.Priority < 0 {
			return fmt.Errorf("rule priority must be non-negative: %w", ErrInvalidScoringRuleSet)
		}
		if rule.UnitKey != "" {
			unit, ok := UnitDefinitionByKey(rule.UnitKey)
			if !ok || unit.ActivityID != rule.ActivityID {
				return fmt.Errorf("unit key %q is not valid for activity %d: %w", rule.UnitKey, rule.ActivityID, ErrInvalidScoringRuleSet)
			}
		}
		if len(rule.LanguageCode) > 10 {
			return fmt.Errorf("language code is too long: %w", ErrInvalidScoringRuleSet)
		}
		if rule.LanguageCode != "" {
			if _, ok := languageCodes[rule.LanguageCode]; !ok {
				return fmt.Errorf("language code %q is not valid: %w", rule.LanguageCode, ErrInvalidScoringRuleSet)
			}
		}
		if rule.Tag != "" {
			tags, err := ValidateAndNormalizeTags([]string{rule.Tag})
			if err != nil || len(tags) != 1 {
				return fmt.Errorf("rule tag is invalid: %w", ErrInvalidScoringRuleSet)
			}
			rule.Tag = tags[0]
		}
	}
	if err := validateScoringRules(req.Rules); err != nil {
		return err
	}
	return nil
}

func (s *ScoringRuleSetManagement) authorizeRuleSetChange(
	ctx context.Context,
	ruleSet *ScoringRuleSet,
) error {
	switch ruleSet.Scope {
	case ScoringRuleSetScopePlatform:
		if !isAdmin(ctx) {
			return ErrForbidden
		}
		return nil
	case ScoringRuleSetScopeContest:
		if ruleSet.ContestID == nil {
			return ErrInvalidScoringRuleSet
		}
		return s.requireContestOwnerBeforeStart(ctx, *ruleSet.ContestID)
	default:
		return ErrInvalidScoringRuleSet
	}
}

func (s *ScoringRuleSetManagement) requireContestOwnerBeforeStart(
	ctx context.Context,
	contestID uuid.UUID,
) error {
	contest, err := s.requireContestOwner(ctx, contestID)
	if err != nil {
		return err
	}
	if !s.clock.Now().Before(contest.ContestStart) {
		return fmt.Errorf("contest scoring cannot change after the contest starts: %w", ErrConflict)
	}
	return nil
}

func (s *ScoringRuleSetManagement) requireContestOwner(
	ctx context.Context,
	contestID uuid.UUID,
) (*ContestView, error) {
	contest, err := s.repo.FindContestByID(ctx, &ContestFindRequest{ID: contestID})
	if err != nil {
		return nil, err
	}
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	if !isAdmin(ctx) && contest.OwnerUserID.String() != session.Subject {
		return nil, ErrForbidden
	}
	return contest, nil
}
