package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
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
	if p.ActivityID == 0 {
		return nil, errx.NewInvalidInputError("activity_id is required")
	}
	if p.LanguageCode == "" {
		return nil, errx.NewInvalidInputError("language_code is required")
	}
	p.Tags, err = scoring.NormalizeTags(p.Tags)
	if err != nil {
		return nil, err
	}

	valid := map[uuid.UUID]contests.Registration{}
	if len(p.RegistrationIDs) > 0 {
		registrations, err := a.contests.ListOngoingRegistrations(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, registration := range registrations.Registrations {
			valid[registration.ID] = registration
		}
		for _, id := range p.RegistrationIDs {
			registration, ok := valid[id]
			if !ok {
				return nil, errx.NewInvalidInputError("registration_id is not ongoing for the current user")
			}
			if !registrationAllowsScoring(registration, p.LanguageCode, p.ActivityID) {
				return nil, errx.NewInvalidInputError("language_code or activity_id is not allowed for registration_id")
			}
		}
	}
	tracking, err := a.logs.ResolveTracking(ctx, p.ActivityID, p.LanguageCode, p.UnitID, p.UnitKey, p.Amount, p.DurationSeconds)
	if err != nil {
		return nil, err
	}
	tracking, contestTrackings, err := a.scoreLog(ctx, "create", tracking, p.ActivityID, p.LanguageCode, p.UnitID, p.UnitKey, p.Amount, p.DurationSeconds, p.Tags, p.RegistrationIDs, valid)
	if err != nil {
		return nil, err
	}

	eligibleOfficial := false
	for _, id := range p.RegistrationIDs {
		if valid[id].Contest.Official {
			eligibleOfficial = true
		}
	}
	mutation := logs.Mutation{
		UserID:                      userID,
		LanguageCode:                p.LanguageCode,
		ActivityID:                  p.ActivityID,
		Description:                 p.Description,
		Tags:                        p.Tags,
		Tracking:                    tracking,
		ContestTrackings:            contestTrackings,
		EligibleOfficialLeaderboard: eligibleOfficial,
		Year:                        int16(now.Year()),
		Now:                         now,
	}
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, userID); err != nil {
			return err
		}
		mutation.ID = uuid.New()
		if err := a.logs.Create(ctx, mutation); err != nil {
			return err
		}
		for _, ct := range mutation.ContestTrackings {
			if err := a.logs.CreateContest(ctx, mutation.ID, ct); err != nil {
				return err
			}
		}
		for _, tag := range mutation.Tags {
			if err := a.logs.InsertTag(ctx, mutation.ID, userID, tag); err != nil {
				return err
			}
		}
		seen := map[uuid.UUID]struct{}{}
		for _, ct := range mutation.ContestTrackings {
			if _, ok := seen[ct.ContestID]; ok {
				continue
			}
			seen[ct.ContestID] = struct{}{}
			id := ct.ContestID
			if err := a.logs.InsertOutbox(ctx, userID, &id, nil, "refresh_contest_score"); err != nil {
				return err
			}
		}
		if eligibleOfficial {
			year := mutation.Year
			if err := a.logs.InsertOutbox(ctx, userID, nil, &year, "refresh_official_scores"); err != nil {
				return err
			}
		}
		return nil
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
	p.Tags, err = scoring.NormalizeTags(p.Tags)
	if err != nil {
		return nil, err
	}
	tracking, err := a.logs.ResolveTracking(ctx, existing.Activity.ID, existing.LanguageCode, p.UnitID, p.UnitKey, p.Amount, p.DurationSeconds)
	if err != nil {
		return nil, err
	}
	now := timex.Now()
	scoringEnabled := a.logs.ScoringEngineEnabled()
	ids := make([]uuid.UUID, 0, len(existing.Registrations))
	valid := map[uuid.UUID]contests.Registration{}
	for _, ref := range existing.Registrations {
		if !scoringEnabled || ref.ContestEnd.Before(now) {
			continue
		}
		ids = append(ids, ref.RegistrationID)
		valid[ref.RegistrationID] = contests.Registration{
			ID:        ref.RegistrationID,
			ContestID: ref.ContestID,
		}
	}
	tracking, contestTrackings, err := a.scoreLog(ctx, "update", tracking, existing.Activity.ID, existing.LanguageCode, p.UnitID, p.UnitKey, p.Amount, p.DurationSeconds, p.Tags, ids, valid)
	if err != nil {
		return nil, err
	}
	mutation := logs.Mutation{
		ID:               p.ID,
		UserID:           existing.UserID,
		Description:      p.Description,
		Tags:             p.Tags,
		Tracking:         tracking,
		ContestTrackings: contestTrackings,
		Now:              now,
	}
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, existing.UserID); err != nil {
			return err
		}
		if err := a.logs.LockLog(ctx, p.ID); err != nil {
			return err
		}
		outbox, err := a.logs.OutboxContext(ctx, p.ID)
		if err != nil {
			return err
		}
		if err := a.logs.Update(ctx, mutation); err != nil {
			return err
		}
		if len(contestTrackings) == 0 {
			inherited := tracking
			inherited.RuleSetID = nil
			inherited.RuleIDs = nil
			inherited.Rates = nil
			inherited.Source = ""
			if err := a.logs.UpdateOngoingContests(ctx, p.ID, inherited, now); err != nil {
				return err
			}
		} else {
			for _, ct := range contestTrackings {
				if err := a.logs.UpdateContest(ctx, p.ID, ct, now); err != nil {
					return err
				}
			}
		}
		if err := a.logs.DeleteTags(ctx, p.ID); err != nil {
			return err
		}
		for _, tag := range p.Tags {
			if err := a.logs.InsertTag(ctx, p.ID, existing.UserID, tag); err != nil {
				return err
			}
		}
		contestIDs, err := a.logs.OngoingContestIDs(ctx, p.ID, now)
		if err != nil {
			return err
		}
		for _, id := range contestIDs {
			id := id
			if err := a.logs.InsertOutbox(ctx, outbox.UserID, &id, nil, "refresh_contest_score"); err != nil {
				return err
			}
		}
		if outbox.EligibleOfficial {
			year := outbox.Year
			if err := a.logs.InsertOutbox(ctx, outbox.UserID, nil, &year, "refresh_official_scores"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return a.logs.FindLog(ctx, p.ID, false)
}

func (a *Application) scoreLog(ctx context.Context, operation string, base logs.Tracking, activityID int32, language string, unitID *uuid.UUID, unitKey *string, amount *float32, duration *int32, tags []string, registrationIDs []uuid.UUID, registrations map[uuid.UUID]contests.Registration) (logs.Tracking, []logs.ContestTracking, error) {
	parameters := scoring.PreviewParameters{
		UnitID:          unitID,
		UnitKey:         unitKey,
		ActivityID:      activityID,
		LanguageCode:    language,
		Amount:          amount,
		DurationSeconds: duration,
		Tags:            tags,
	}
	platform, matched, err := a.scoring.ScorePlatform(ctx, parameters)
	mode := "shadow"
	if a.logs.ScoringEngineEnabled() {
		mode = "authoritative"
	}
	comparison := observability.ScoringComparison{
		Operation:    operation,
		Mode:         mode,
		ActivityID:   activityID,
		UnitKey:      base.UnitKey,
		LanguageCode: language,
		LegacyScore:  base.Score,
		Matched:      matched,
	}
	if amount != nil {
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
		contest := make([]logs.ContestTracking, len(registrationIDs))
		for i, id := range registrationIDs {
			contest[i] = logs.ContestTracking{
				RegistrationID: id,
				ContestID:      registrations[id].ContestID,
				Tracking:       base,
			}
		}
		return base, contest, nil
	}
	base = trackingFromEstimate(base, platform)
	contest := make([]logs.ContestTracking, len(registrationIDs))
	for i, id := range registrationIDs {
		registration := registrations[id]
		estimate, err := a.scoring.ScoreContest(ctx, parameters, registration.ContestID, platform)
		if err != nil {
			return logs.Tracking{}, nil, err
		}
		contest[i] = logs.ContestTracking{
			RegistrationID: id,
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
