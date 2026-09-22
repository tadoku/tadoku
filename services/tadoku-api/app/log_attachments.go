package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/scoring"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func (a *Application) UpdateLogContestRegistrations(ctx context.Context, logID uuid.UUID, registrationIDs []uuid.UUID) (*Log, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	callerID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}
	log, err := a.logs.FindLog(ctx, logID, false)
	if err != nil {
		return nil, err
	}
	if log.UserID != callerID {
		return nil, errx.NewForbiddenError("forbidden")
	}

	registrations, err := a.contests.ListOngoingRegistrations(ctx, callerID)
	if err != nil {
		return nil, err
	}
	valid := make(map[uuid.UUID]contests.Registration, len(registrations.Registrations))
	for _, registration := range registrations.Registrations {
		valid[registration.ID] = registration
	}
	for _, registrationID := range registrationIDs {
		registration, ok := valid[registrationID]
		if !ok {
			return nil, errx.NewInvalidInputError("registration_id is not ongoing for the current user")
		}
		if !registration.IsEligibleForScoring(log.LanguageCode, log.Activity.ID) {
			return nil, errx.NewInvalidInputError("language_code or activity_id is not allowed for registration_id")
		}
	}

	now := timex.Now()
	desired := make(map[uuid.UUID]struct{}, len(registrationIDs))
	for _, registrationID := range registrationIDs {
		desired[registrationID] = struct{}{}
	}
	current := make(map[uuid.UUID]logs.RegistrationReference)
	for _, reference := range log.Registrations {
		if reference.ContestEnd.Add(24 * time.Hour).After(now) {
			current[reference.RegistrationID] = reference
		}
	}

	attachments := make([]logs.ContestTracking, 0)
	for _, registrationID := range registrationIDs {
		if _, exists := current[registrationID]; exists {
			continue
		}
		registration := valid[registrationID]
		attachments = append(attachments, logs.ContestTracking{
			RegistrationID: registrationID,
			ContestID:      registration.ContestID,
			Tracking:       log.Tracking,
		})
	}
	detachments := make([]uuid.UUID, 0)
	for registrationID, reference := range current {
		if _, exists := desired[registrationID]; !exists {
			detachments = append(detachments, reference.ContestID)
		}
	}
	if len(attachments) == 0 && len(detachments) == 0 {
		return log, nil
	}
	if len(attachments) > 0 && log.Tracking.Amount == nil && log.Tracking.DurationSeconds == nil {
		return nil, errx.NewInvalidInputError("log tracking data is required for contest attachment")
	}

	if a.logs.ScoringEngineEnabled() {
		unitKey := log.Tracking.UnitKey
		parameters := scoring.PreviewParameters{
			UnitKey:         &unitKey,
			ActivityID:      log.Activity.ID,
			LanguageCode:    log.LanguageCode,
			Amount:          log.Tracking.Amount,
			DurationSeconds: log.Tracking.DurationSeconds,
			Tags:            log.Tags,
		}
		for i := range attachments {
			estimate, err := a.scoring.ScoreResolvedContest(ctx, parameters, attachments[i].ContestID)
			if err != nil {
				return nil, err
			}
			if estimate != nil {
				attachments[i].Tracking = trackingFromEstimate(log.Tracking, *estimate)
			}
		}
	}

	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, log.UserID); err != nil {
			return err
		}
		if err := a.logs.LockLog(ctx, logID); err != nil {
			return err
		}
		before, err := a.logs.OutboxContext(ctx, logID)
		if err != nil {
			return err
		}
		for _, contestID := range detachments {
			if err := a.logs.DetachContest(ctx, logID, contestID); err != nil {
				return err
			}
		}
		for _, attachment := range attachments {
			if err := a.logs.CreateContest(ctx, logID, attachment); err != nil {
				return err
			}
		}
		if err := a.logs.RecomputeOfficialEligibility(ctx, logID, now); err != nil {
			return err
		}
		after, err := a.logs.OutboxContext(ctx, logID)
		if err != nil {
			return err
		}

		affected := make(map[uuid.UUID]struct{}, len(detachments)+len(attachments))
		for _, contestID := range detachments {
			affected[contestID] = struct{}{}
		}
		for _, attachment := range attachments {
			affected[attachment.ContestID] = struct{}{}
		}
		for contestID := range affected {
			contestID := contestID
			if err := a.logs.InsertOutbox(ctx, log.UserID, &contestID, nil, "refresh_contest_score"); err != nil {
				return err
			}
		}
		if before.EligibleOfficial || after.EligibleOfficial {
			year := before.Year
			if err := a.logs.InsertOutbox(ctx, log.UserID, nil, &year, "refresh_official_scores"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return a.logs.FindLog(ctx, logID, false)
}

func (a *Application) DeleteLog(ctx context.Context, logID uuid.UUID) error {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}

	callerID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return errx.NewUnauthorizedError("unauthorized")
	}
	log, err := a.logs.FindLog(ctx, logID, false)
	if err != nil {
		return err
	}
	if log.UserID != callerID {
		admin, err := a.permissions.IsAdmin(ctx)
		if err != nil {
			return err
		}
		if !admin {
			return errx.NewForbiddenError("forbidden")
		}
	}

	now := timex.Now()
	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, log.UserID); err != nil {
			return err
		}
		if err := a.logs.LockLog(ctx, logID); err != nil {
			return err
		}
		allowed, err := a.logs.CanDelete(ctx, logID, now)
		if err != nil {
			return err
		}
		if !allowed {
			return errx.NewForbiddenError("forbidden")
		}
		outbox, err := a.logs.OutboxContext(ctx, logID)
		if err != nil {
			return err
		}
		contestIDs, err := a.logs.AttachedContestIDs(ctx, logID)
		if err != nil {
			return err
		}
		if err := a.logs.SoftDelete(ctx, logID, now); err != nil {
			return err
		}
		for _, contestID := range contestIDs {
			contestID := contestID
			if err := a.logs.InsertOutbox(ctx, log.UserID, &contestID, nil, "refresh_contest_score"); err != nil {
				return err
			}
		}
		if outbox.EligibleOfficial {
			year := outbox.Year
			if err := a.logs.InsertOutbox(ctx, log.UserID, nil, &year, "refresh_official_scores"); err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *Application) DetachContestLog(ctx context.Context, contestID, logID uuid.UUID, reason string) error {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}

	callerID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return errx.NewUnauthorizedError("unauthorized")
	}
	contest, err := a.contests.FindContestByID(ctx, contestID, false)
	if err != nil {
		return err
	}
	if callerID != contest.OwnerUserID {
		admin, err := a.permissions.IsAdmin(ctx)
		if err != nil {
			return err
		}
		if !admin {
			return errx.NewForbiddenError("forbidden")
		}
	}
	if _, err := a.logs.FindLog(ctx, logID, false); err != nil {
		return err
	}

	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		outbox, err := a.logs.OutboxContext(ctx, logID)
		if err != nil {
			return err
		}
		if err := a.profile.LockUser(ctx, outbox.UserID); err != nil {
			return err
		}
		if err := a.audit.Record(ctx, audit.Event{
			ActorID: callerID,
			Action:  "detach_log",
			Metadata: map[string]any{
				"contest_id": contestID.String(),
				"log_id":     logID.String(),
			},
			Description: reason,
		}); err != nil {
			return err
		}
		if err := a.logs.DetachContest(ctx, logID, contestID); err != nil {
			return err
		}
		if err := a.logs.InsertOutbox(ctx, outbox.UserID, &contestID, nil, "refresh_contest_score"); err != nil {
			return err
		}
		if outbox.EligibleOfficial {
			year := outbox.Year
			if err := a.logs.InsertOutbox(ctx, outbox.UserID, nil, &year, "refresh_official_scores"); err != nil {
				return err
			}
		}
		return nil
	})
}
