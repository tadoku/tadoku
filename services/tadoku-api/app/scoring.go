package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type ScorePreviewParameters struct {
	RegistrationIDs []uuid.UUID
	UnitID          *uuid.UUID
	UnitKey         *string
	ActivityID      int32
	LanguageCode    string
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
}

type ScorePreview = scoring.Preview
type ScoreEstimate = scoring.Estimate
type ScoringRuleSet = scoring.RuleSet
type ScoringRule = scoring.Rule
type ScoringSource = scoring.Source

type PlatformScoringRuleSetDraftParameters struct {
	Rules []ScoringRule
}

type ContestScoringRuleSetDraftParameters struct {
	Mode              string
	FallbackRuleSetID *uuid.UUID
	Rules             []ScoringRule
}

func (a *Application) PreviewScore(ctx context.Context, parameters ScorePreviewParameters) (*ScorePreview, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	userID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}

	featureParameters := scoring.PreviewParameters{
		UnitID:          parameters.UnitID,
		UnitKey:         parameters.UnitKey,
		ActivityID:      parameters.ActivityID,
		LanguageCode:    parameters.LanguageCode,
		Amount:          parameters.Amount,
		DurationSeconds: parameters.DurationSeconds,
		Tags:            parameters.Tags,
		Contests:        make([]scoring.PreviewContest, 0, len(parameters.RegistrationIDs)),
	}
	if err := scoring.ValidatePreview(featureParameters); err != nil {
		return nil, err
	}
	if len(parameters.RegistrationIDs) == 0 {
		return a.scoring.Preview(ctx, featureParameters)
	}

	selected, err := a.contests.SelectRegistrationsForScoring(
		ctx,
		userID,
		parameters.RegistrationIDs,
		parameters.LanguageCode,
		parameters.ActivityID,
	)
	if err != nil {
		return nil, err
	}

	for _, registration := range selected {
		featureParameters.Contests = append(featureParameters.Contests, scoring.PreviewContest{
			RegistrationID: registration.RegistrationID,
			ContestID:      registration.ContestID,
		})
	}

	return a.scoring.Preview(ctx, featureParameters)
}

func (a *Application) ListPlatformScoringRuleSets(ctx context.Context) ([]ScoringRuleSet, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return a.scoring.ListPlatformRuleSets(ctx, a.permissions.IsAdminOrFalse(ctx))
}

func (a *Application) ListContestScoringRuleSets(ctx context.Context, contestID uuid.UUID) ([]ScoringRuleSet, error) {
	caller := identity.FromContext(ctx)
	if caller == nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}

	contest, err := a.contests.FindContestByID(ctx, contestID, false)
	if err != nil {
		return nil, err
	}
	callerID, parseErr := uuid.Parse(caller.Subject)
	if (parseErr != nil || callerID != contest.OwnerUserID) && !a.permissions.IsAdminOrFalse(ctx) {
		return nil, errx.NewForbiddenError("forbidden")
	}
	return a.scoring.ListContestRuleSets(ctx, contestID)
}

func (a *Application) CreatePlatformScoringRuleSetDraft(ctx context.Context, parameters PlatformScoringRuleSetDraftParameters) (*ScoringRuleSet, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	draft := scoring.PlatformDraftParameters{
		DraftRules: scoring.DraftRules{
			Rules: parameters.Rules,
		},
	}
	if err := a.normalizeScoringDraftRules(ctx, &draft.DraftRules); err != nil {
		return nil, err
	}
	draft.CreatedAt = timex.Now()

	var created *ScoringRuleSet
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var createErr error
		created, createErr = a.scoring.CreatePlatformDraft(ctx, draft)
		return createErr
	})

	return created, err
}

func (a *Application) CreateContestScoringRuleSetDraft(ctx context.Context, contestID uuid.UUID, parameters ContestScoringRuleSetDraftParameters) (*ScoringRuleSet, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	contest, err := a.contests.FindContestByID(ctx, contestID, false)
	if err != nil {
		return nil, err
	}
	caller := identity.FromContext(ctx)
	if caller.Subject != contest.OwnerUserID.String() {
		if err := a.permissions.RequireAdmin(ctx); err != nil {
			return nil, err
		}
	}
	if !timex.Now().Before(contest.ContestStart) {
		return nil, errx.NewConflictError("contest scoring cannot change after the contest starts")
	}

	configuration, err := scoring.NewContestDraftConfiguration(parameters.Mode, parameters.FallbackRuleSetID)
	if err != nil {
		return nil, err
	}

	draft := scoring.ContestDraftParameters{
		ContestID:     contestID,
		Configuration: configuration,
		DraftRules: scoring.DraftRules{
			Rules: parameters.Rules,
		},
	}
	if err := a.scoring.ValidateContestDraftConfiguration(ctx, &draft); err != nil {
		return nil, err
	}
	if err := a.normalizeScoringDraftRules(ctx, &draft.DraftRules); err != nil {
		return nil, err
	}
	draft.CreatedAt = timex.Now()

	var created *ScoringRuleSet
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var createErr error
		created, createErr = a.scoring.CreateContestDraft(ctx, draft)
		return createErr
	})

	return created, err
}

func (a *Application) normalizeScoringDraftRules(ctx context.Context, rules *scoring.DraftRules) error {
	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return err
	}

	return a.scoring.NormalizeDraftRules(rules, languages)
}

func (a *Application) PublishScoringRuleSet(ctx context.Context, id uuid.UUID) (*ScoringRuleSet, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	ruleSet, err := a.scoring.FindRuleSet(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := a.authorizeScoringRuleSetChange(ctx, *ruleSet); err != nil {
		return nil, err
	}

	var published *ScoringRuleSet
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		published, err = a.scoring.PublishRuleSet(ctx, *ruleSet, timex.Now())
		return err
	})
	return published, err
}

func (a *Application) ActivateScoringRuleSet(ctx context.Context, id uuid.UUID) error {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}

	ruleSet, err := a.scoring.FindRuleSet(ctx, id)
	if err != nil {
		return err
	}
	if err := a.authorizeScoringRuleSetChange(ctx, *ruleSet); err != nil {
		return err
	}

	return a.scoring.ActivateRuleSet(ctx, *ruleSet, timex.Now())
}

func (a *Application) authorizeScoringRuleSetChange(ctx context.Context, ruleSet ScoringRuleSet) error {
	switch ruleSet.Scope {
	case "platform":
		return a.permissions.RequireAdmin(ctx)
	case "contest":
		if ruleSet.ContestID == nil {
			return errx.NewInvalidInputError("contest scoring rule set requires contest_id")
		}
		contest, err := a.contests.FindContestByID(ctx, *ruleSet.ContestID, false)
		if err != nil {
			return err
		}
		if identity.FromContext(ctx).Subject != contest.OwnerUserID.String() {
			if err := a.permissions.RequireAdmin(ctx); err != nil {
				return err
			}
		}
		if !timex.Now().Before(contest.ContestStart) {
			return errx.NewConflictError("contest scoring cannot change after the contest starts")
		}
		return nil
	default:
		return errx.NewInvalidInputError("scoring rule set scope is invalid")
	}
}
