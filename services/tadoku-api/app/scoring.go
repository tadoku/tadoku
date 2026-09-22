package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
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

func (a *Application) PreviewScore(ctx context.Context, parameters ScorePreviewParameters) (*ScorePreview, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	userID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}

	featureParameters, err := scoring.PreparePreview(scoring.PreviewParameters{
		UnitID:          parameters.UnitID,
		UnitKey:         parameters.UnitKey,
		ActivityID:      parameters.ActivityID,
		LanguageCode:    parameters.LanguageCode,
		Amount:          parameters.Amount,
		DurationSeconds: parameters.DurationSeconds,
		Tags:            parameters.Tags,
		Contests:        make([]scoring.PreviewContest, 0, len(parameters.RegistrationIDs)),
	})
	if err != nil {
		return nil, err
	}
	if len(parameters.RegistrationIDs) == 0 {
		return a.scoring.Preview(ctx, featureParameters)
	}

	registrations, err := a.contests.ListOngoingRegistrations(ctx, userID)
	if err != nil {
		return nil, err
	}
	selected, err := contests.SelectRegistrationsForScoring(
		parameters.RegistrationIDs,
		registrations.Registrations,
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
	contest, err := a.contests.FindContestByID(ctx, contestID, false)
	if err != nil {
		return nil, err
	}
	caller := identity.FromContext(ctx)
	if caller == nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}
	callerID, parseErr := uuid.Parse(caller.Subject)
	if (parseErr != nil || callerID != contest.OwnerUserID) && !a.permissions.IsAdminOrFalse(ctx) {
		return nil, errx.NewForbiddenError("forbidden")
	}
	return a.scoring.ListContestRuleSets(ctx, contestID)
}
