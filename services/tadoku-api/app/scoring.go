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

type ScoringRuleSetDraftParameters struct {
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
			RegistrationID: registration.ID,
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

func (a *Application) CreatePlatformScoringRuleSetDraft(ctx context.Context, parameters ScoringRuleSetDraftParameters) (*ScoringRuleSet, error) {
	admin, err := a.permissions.IsAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !admin {
		return nil, errx.NewForbiddenError("forbidden")
	}

	parameters.Mode = ""
	parameters.FallbackRuleSetID = nil
	return a.createScoringRuleSetDraft(ctx, "platform", nil, parameters)
}

func (a *Application) CreateContestScoringRuleSetDraft(ctx context.Context, contestID uuid.UUID, parameters ScoringRuleSetDraftParameters) (*ScoringRuleSet, error) {
	contest, err := a.contests.FindContestByID(ctx, contestID, false)
	if err != nil {
		return nil, err
	}
	caller := identity.FromContext(ctx)
	if caller == nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}
	if caller.Subject == contest.OwnerUserID.String() {
		if err := a.permissions.RequireAuthenticated(ctx); err != nil {
			return nil, err
		}
	} else {
		admin, err := a.permissions.IsAdmin(ctx)
		if err != nil {
			return nil, err
		}
		if !admin {
			return nil, errx.NewForbiddenError("forbidden")
		}
	}
	if !timex.Now().Before(contest.ContestStart) {
		return nil, errx.NewConflictError("contest scoring cannot change after the contest starts")
	}

	return a.createScoringRuleSetDraft(ctx, "contest", &contestID, parameters)
}

func (a *Application) createScoringRuleSetDraft(ctx context.Context, scope string, contestID *uuid.UUID, parameters ScoringRuleSetDraftParameters) (*ScoringRuleSet, error) {
	var mode scoring.Mode
	if scope == "contest" {
		var err error
		mode, err = scoring.ParseMode(parameters.Mode)
		if err != nil {
			return nil, err
		}
	}

	draft := scoring.DraftParameters{
		ContestID:         contestID,
		Mode:              mode,
		FallbackRuleSetID: parameters.FallbackRuleSetID,
		Rules:             parameters.Rules,
	}
	if err := a.scoring.ValidateDraftConfiguration(ctx, scope, &draft); err != nil {
		return nil, err
	}

	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}
	languageCodes := make(map[string]struct{}, len(languages))
	for _, language := range languages {
		languageCodes[language.Code] = struct{}{}
	}

	draft.LanguageCodes = languageCodes
	if err := a.scoring.ValidateDraftRules(&draft); err != nil {
		return nil, err
	}
	draft.CreatedAt = timex.Now()

	var created *ScoringRuleSet
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var createErr error
		created, createErr = a.scoring.CreateDraft(ctx, scope, draft)
		return createErr
	})
	return created, err
}
