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
	if parameters.ActivityID == 0 {
		return nil, errx.NewInvalidInputError("activity_id is required")
	}
	if parameters.LanguageCode == "" {
		return nil, errx.NewInvalidInputError("language_code is required")
	}
	tags, err := scoring.NormalizeTags(parameters.Tags)
	if err != nil {
		return nil, err
	}

	featureParameters := scoring.PreviewParameters{
		UnitID:          parameters.UnitID,
		UnitKey:         parameters.UnitKey,
		ActivityID:      parameters.ActivityID,
		LanguageCode:    parameters.LanguageCode,
		Amount:          parameters.Amount,
		DurationSeconds: parameters.DurationSeconds,
		Tags:            tags,
		Contests:        make([]scoring.PreviewContest, 0, len(parameters.RegistrationIDs)),
	}
	if len(parameters.RegistrationIDs) == 0 {
		return a.scoring.Preview(ctx, featureParameters)
	}

	registrations, err := a.contests.ListOngoingRegistrations(ctx, userID)
	if err != nil {
		return nil, err
	}
	available := make(map[uuid.UUID]contests.Registration, len(registrations.Registrations))
	for _, registration := range registrations.Registrations {
		available[registration.ID] = registration
	}
	for _, registrationID := range parameters.RegistrationIDs {
		registration, ok := available[registrationID]
		if !ok {
			return nil, errx.NewInvalidInputError("registration_id is not ongoing for the current user")
		}
		if !registrationAllowsScoring(registration, parameters.LanguageCode, parameters.ActivityID) {
			return nil, errx.NewInvalidInputError("language_code or activity_id is not allowed for registration_id")
		}
		featureParameters.Contests = append(featureParameters.Contests, scoring.PreviewContest{
			RegistrationID: registrationID,
			ContestID:      registration.ContestID,
		})
	}
	return a.scoring.Preview(ctx, featureParameters)
}

func registrationAllowsScoring(registration contests.Registration, languageCode string, activityID int32) bool {
	languageAllowed := false
	for _, language := range registration.LanguageCodes {
		if language == languageCode {
			languageAllowed = true
			break
		}
	}
	if !languageAllowed || registration.Contest == nil {
		return false
	}
	for _, activity := range registration.Contest.AllowedActivities {
		if activity.ID == activityID {
			return true
		}
	}
	return false
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
