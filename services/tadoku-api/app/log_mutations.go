package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/observability"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type LogCreateParameters struct {
	RegistrationIDs []uuid.UUID
	UnitID          *uuid.UUID
	UnitKey         *string
	ActivityID      int32
	LanguageCode    string
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
	Description     *string
}

type LogUpdateParameters struct {
	ID              uuid.UUID
	UnitID          *uuid.UUID
	UnitKey         *string
	Amount          *float32
	DurationSeconds *int32
	Tags            []string
	Description     *string
}

func (a *Application) CreateLog(ctx context.Context, p LogCreateParameters) (*Log, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	caller := identity.FromContext(ctx)
	now := timex.Now()
	userID, err := a.profile.SynchronizeUser(ctx, caller, now)
	if err != nil {
		return nil, err
	}
	scored, err := a.scoreLog(ctx, userID, p)
	if err != nil {
		return nil, err
	}

	mutation := logs.Mutation{
		ID:                          uuid.New(),
		UserID:                      userID,
		LanguageCode:                p.LanguageCode,
		ActivityID:                  p.ActivityID,
		Description:                 p.Description,
		Tags:                        scored.tags,
		Tracking:                    scored.tracking,
		ContestTrackings:            scored.contestTrackings,
		EligibleOfficialLeaderboard: scored.eligibleOfficial,
		Year:                        int16(now.Year()),
		Now:                         now,
	}
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, userID); err != nil {
			return err
		}
		return a.logs.Create(ctx, mutation)
	})
	if err != nil {
		return nil, err
	}
	return a.logs.FindLog(ctx, mutation.ID, false)
}

func (a *Application) UpdateLog(ctx context.Context, p LogUpdateParameters) (*Log, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	callerID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}
	existing, err := a.logs.FindLog(ctx, p.ID, false)
	if err != nil {
		return nil, err
	}
	if callerID != existing.UserID {
		admin, err := a.permissions.IsAdmin(ctx)
		if err != nil {
			return nil, err
		}
		if !admin {
			return nil, errx.NewForbiddenError("forbidden")
		}
	}
	scored, now, err := a.scoreUpdatedLog(ctx, existing, p)
	if err != nil {
		return nil, err
	}
	mutation := logs.Mutation{
		ID:               p.ID,
		UserID:           existing.UserID,
		Description:      p.Description,
		Tags:             scored.tags,
		Tracking:         scored.tracking,
		ContestTrackings: scored.contestTrackings,
		Now:              now,
	}
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, existing.UserID); err != nil {
			return err
		}
		return a.logs.Update(ctx, mutation)
	})
	if err != nil {
		return nil, err
	}
	return a.logs.FindLog(ctx, p.ID, false)
}

type scoredLog struct {
	tracking         logs.Tracking
	contestTrackings []logs.ContestTracking
	tags             []string
	eligibleOfficial bool
}

func (a *Application) scoreLog(ctx context.Context, userID uuid.UUID, p LogCreateParameters) (scoredLog, error) {
	if p.ActivityID == 0 {
		return scoredLog{}, errx.NewInvalidInputError("activity_id is required")
	}
	if p.LanguageCode == "" {
		return scoredLog{}, errx.NewInvalidInputError("language_code is required")
	}
	tags, err := scoring.NormalizeTags(p.Tags)
	if err != nil {
		return scoredLog{}, err
	}

	parameters := scoring.PreviewParameters{
		UnitID:          p.UnitID,
		UnitKey:         p.UnitKey,
		ActivityID:      p.ActivityID,
		LanguageCode:    p.LanguageCode,
		Amount:          p.Amount,
		DurationSeconds: p.DurationSeconds,
		Tags:            tags,
	}
	eligibleOfficial := false
	if len(p.RegistrationIDs) > 0 {
		selected, err := a.contests.SelectRegistrationsForScoring(ctx, userID, p.RegistrationIDs, p.LanguageCode, p.ActivityID)
		if err != nil {
			return scoredLog{}, err
		}
		for _, registration := range selected {
			parameters.Contests = append(parameters.Contests, scoring.PreviewContest{
				RegistrationID: registration.ID,
				ContestID:      registration.ContestID,
			})
			eligibleOfficial = eligibleOfficial || registration.Contest.Official
		}
	}

	base, err := a.logs.ResolveTracking(ctx, p.ActivityID, p.LanguageCode, p.UnitID, p.UnitKey, p.Amount, p.DurationSeconds)
	if err != nil {
		return scoredLog{}, err
	}
	tracking, contestTrackings, err := a.scoreTracking(ctx, "create", base, parameters)
	if err != nil {
		return scoredLog{}, err
	}
	return scoredLog{tracking: tracking, contestTrackings: contestTrackings, tags: tags, eligibleOfficial: eligibleOfficial}, nil
}

func (a *Application) scoreUpdatedLog(ctx context.Context, existing *logs.Log, p LogUpdateParameters) (scoredLog, time.Time, error) {
	tags, err := scoring.NormalizeTags(p.Tags)
	if err != nil {
		return scoredLog{}, time.Time{}, err
	}
	base, err := a.logs.ResolveTracking(ctx, existing.Activity.ID, existing.LanguageCode, p.UnitID, p.UnitKey, p.Amount, p.DurationSeconds)
	if err != nil {
		return scoredLog{}, time.Time{}, err
	}
	now := timex.Now()
	parameters := scoring.PreviewParameters{
		UnitID:          p.UnitID,
		UnitKey:         p.UnitKey,
		ActivityID:      existing.Activity.ID,
		LanguageCode:    existing.LanguageCode,
		Amount:          p.Amount,
		DurationSeconds: p.DurationSeconds,
		Tags:            tags,
	}
	for _, registration := range a.logs.RegistrationsForRescoring(existing, now) {
		parameters.Contests = append(parameters.Contests, scoring.PreviewContest{
			RegistrationID: registration.RegistrationID,
			ContestID:      registration.ContestID,
		})
	}
	tracking, contestTrackings, err := a.scoreTracking(ctx, "update", base, parameters)
	if err != nil {
		return scoredLog{}, time.Time{}, err
	}
	return scoredLog{tracking: tracking, contestTrackings: contestTrackings, tags: tags}, now, nil
}

func (a *Application) scoreTracking(ctx context.Context, operation string, base logs.Tracking, parameters scoring.PreviewParameters) (logs.Tracking, []logs.ContestTracking, error) {
	platform, matched, err := a.scoring.ScorePlatform(ctx, parameters)
	mode := "shadow"
	if a.logs.ScoringEngineEnabled() {
		mode = "authoritative"
	}
	comparison := observability.ScoringComparison{
		Operation:    operation,
		Mode:         mode,
		ActivityID:   parameters.ActivityID,
		UnitKey:      base.UnitKey,
		LanguageCode: parameters.LanguageCode,
		LegacyScore:  base.Score,
		Matched:      matched,
	}
	if parameters.Amount != nil {
		comparison.ScoreSource = "amount"
	} else {
		comparison.ScoreSource = "duration_minutes"
	}
	if err != nil {
		comparison.ErrorType = scoringErrorType(err)
	} else {
		comparison.EngineScore = &platform.Score
		comparison.RuleSetID = platform.RuleSetID
		for _, rule := range platform.Rules {
			comparison.AppliedRuleIDs = append(comparison.AppliedRuleIDs, rule.RuleID)
		}
	}
	a.scoringObserver.Observe(ctx, comparison)
	if err != nil {
		if a.logs.ScoringEngineEnabled() {
			return logs.Tracking{}, nil, err
		}
	}
	if !a.logs.ScoringEngineEnabled() {
		contest := make([]logs.ContestTracking, len(parameters.Contests))
		for i, registration := range parameters.Contests {
			contest[i] = logs.ContestTracking{
				RegistrationID: registration.RegistrationID,
				ContestID:      registration.ContestID,
				Tracking:       base,
			}
		}
		return base, contest, nil
	}
	base = trackingFromEstimate(base, platform)
	contest := make([]logs.ContestTracking, len(parameters.Contests))
	for i, registration := range parameters.Contests {
		estimate, err := a.scoring.ScoreContest(ctx, parameters, registration.ContestID, platform)
		if err != nil {
			return logs.Tracking{}, nil, err
		}
		contest[i] = logs.ContestTracking{
			RegistrationID: registration.RegistrationID,
			ContestID:      registration.ContestID,
			Tracking:       trackingFromEstimate(base, estimate),
		}
	}
	return base, contest, nil
}

func scoringErrorType(err error) string {
	if errors.Is(err, scoring.ErrRuleSetNotFound) {
		return "scoring_rule_set_not_found"
	}
	switch errx.KindOf(err) {
	case errx.InvalidInput:
		return "invalid_scoring_input"
	case errx.Internal:
		return "invalid_scoring_rule_set"
	default:
		return "evaluation_failed"
	}
}

func trackingFromEstimate(base logs.Tracking, estimate scoring.Estimate) logs.Tracking {
	base.Score = estimate.Score
	base.RuleSetID = estimate.RuleSetID
	base.Source = string(estimate.Source)
	base.RuleIDs = nil
	base.Rates = nil
	for _, rule := range estimate.Rules {
		base.RuleIDs = append(base.RuleIDs, rule.RuleID)
		base.Rates = append(base.Rates, rule.Rate)
	}
	return base
}
