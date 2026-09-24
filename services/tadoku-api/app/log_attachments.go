package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func (a *Application) UpdateLogContestRegistrations(ctx context.Context, logID uuid.UUID, registrationIDs []uuid.UUID) (*Log, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	log, err := a.logs.FindLog(ctx, logID)
	if err != nil {
		return nil, err
	}
	if err := a.requireOwnerOrAdmin(ctx, log.UserID); err != nil {
		return nil, err
	}

	targets, err := a.contests.SelectRegistrationsForScoring(ctx, log.UserID, registrationIDs, log.LanguageCode, log.Activity.ID)
	if err != nil {
		return nil, err
	}

	now := timex.Now()
	toAttach, detachments, err := a.logs.PlanContestRegistrationUpdate(log, targets, now)
	if err != nil {
		return nil, err
	}
	if len(toAttach) == 0 && len(detachments) == 0 {
		return log, nil
	}

	attachments, err := a.scoring.ScoreLogAttachments(ctx, logscore.Input{
		ActivityID:   log.Activity.ID,
		LanguageCode: log.LanguageCode,
		Tags:         log.Tags,
	}, log.Tracking, toAttach)
	if err != nil {
		return nil, err
	}

	var updated *Log
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, log.UserID); err != nil {
			return err
		}
		if err := a.logs.UpdateContestRegistrations(ctx, logID, now, attachments, detachments); err != nil {
			return err
		}

		var err error
		updated, err = a.logs.FindLog(ctx, logID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (a *Application) DeleteLog(ctx context.Context, logID uuid.UUID) error {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}

	log, err := a.logs.FindLog(ctx, logID)
	if err != nil {
		return err
	}
	if err := a.requireOwnerOrAdmin(ctx, log.UserID); err != nil {
		return err
	}

	now := timex.Now()
	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, log.UserID); err != nil {
			return err
		}
		return a.logs.Delete(ctx, logID, now)
	})
}

func (a *Application) DetachContestLog(ctx context.Context, contestID, logID uuid.UUID, reason string) error {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}

	actorID, err := identity.RequireActorID(ctx)
	if err != nil {
		return err
	}
	contest, err := a.contests.FindContestByID(ctx, contestID, false)
	if err != nil {
		return err
	}
	if err := a.requireOwnerOrAdmin(ctx, contest.OwnerUserID); err != nil {
		return err
	}
	log, err := a.logs.FindLog(ctx, logID)
	if err != nil {
		return err
	}

	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, log.UserID); err != nil {
			return err
		}
		if err := a.audit.Record(ctx, audit.Event{
			ActorID: actorID,
			Action:  "detach_log",
			Metadata: map[string]any{
				"contest_id": contestID.String(),
				"log_id":     logID.String(),
			},
			Description: reason,
		}); err != nil {
			return err
		}
		return a.logs.ModerateDetach(ctx, logID, contestID)
	})
}
