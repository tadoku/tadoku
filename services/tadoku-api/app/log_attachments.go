package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/logscore"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
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

	targets, err := a.contests.SelectRegistrationsForScoring(ctx, callerID, registrationIDs, log.LanguageCode, log.Activity.ID)
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
		updated, err = a.logs.FindLog(ctx, logID, false)
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

	callerID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return errx.NewUnauthorizedError("unauthorized")
	}
	log, err := a.logs.FindLog(ctx, logID, false)
	if err != nil {
		return err
	}
	if log.UserID != callerID {
		if err := a.permissions.RequireAdmin(ctx); err != nil {
			return err
		}
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

	callerID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return errx.NewUnauthorizedError("unauthorized")
	}
	contest, err := a.contests.FindContestByID(ctx, contestID, false)
	if err != nil {
		return err
	}
	if callerID != contest.OwnerUserID {
		if err := a.permissions.RequireAdmin(ctx); err != nil {
			return err
		}
	}
	log, err := a.logs.FindLog(ctx, logID, false)
	if err != nil {
		return err
	}

	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, log.UserID); err != nil {
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
		return a.logs.ModerateDetach(ctx, logID, contestID)
	})
}
