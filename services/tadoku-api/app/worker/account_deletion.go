package worker

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/app/accountdeletion"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type AccountDeletion struct {
	DB            *pgxpool.Pool
	Profile       *profile.Service
	Logs          *logs.Service
	Contests      *contests.Service
	Relationships *permissions.Relationships
	Identity      *identity.Deletion
	Audit         *audit.Service
	ContestEvents *contestevents.Service
}

// DeleteAccount locks the user row before any other row of the deletion, the reverse of the settings writes'
// account-then-user order. That is safe because every account row lock skips deletion-locked users.
func (a *Application) DeleteAccount(ctx context.Context, job jobs.DeleteAccountV1) error {
	deletion := a.accountDeletion
	if deletion == nil || a.email == nil {
		return &permanentError{err: errors.New("worker account deletion is not configured")}
	}

	now := timex.Now()
	err := postgres.RunInTransaction(ctx, deletion.DB, func(ctx context.Context) error {
		tombstoned, err := deletion.Profile.Tombstone(ctx, job.UserID, now)
		if err != nil || !tombstoned {
			return err
		}
		return a.removeAccountData(ctx, job.UserID, now)
	})
	if errors.Is(err, profile.ErrDeletionNotRequested) || errors.Is(err, profile.ErrLocalUserNotFound) {
		return &permanentError{err: err}
	}
	return err
}

func (a *Application) removeAccountData(ctx context.Context, userID uuid.UUID, now time.Time) error {
	deletion := a.accountDeletion
	removedLogs, err := deletion.Logs.RemoveUser(ctx, userID, now)
	if err != nil {
		return err
	}
	removedRegistrations, err := deletion.Contests.RemoveUser(ctx, userID, now)
	if err != nil {
		return err
	}
	if err := deletion.Relationships.RevokeUser(ctx, userID); err != nil {
		return err
	}
	notices, err := deletion.Identity.DeleteAccount(ctx, userID, now)
	if err != nil {
		return err
	}

	followUp := slices.Concat(removedLogs.Jobs, removedRegistrations.Jobs)
	for _, notice := range notices {
		written, err := a.email.Write(ctx, notice)
		if err != nil {
			return err
		}
		followUp = append(followUp, written.Jobs...)
	}

	err = deletion.Audit.Record(ctx, audit.Event{
		ActorID: userID,
		Action:  "complete_account_deletion",
		Metadata: map[string]any{
			"target_user_id":        userID.String(),
			"deleted_logs":          removedLogs.DeletedLogs,
			"kept_logs":             removedLogs.KeptLogs,
			"deleted_registrations": removedRegistrations.DeletedRegistrations,
			"kept_registrations":    removedRegistrations.KeptRegistrations,
			"owned_contests":        removedRegistrations.OwnedContests,
		},
	})
	if err != nil {
		return err
	}
	if err := deletion.ContestEvents.Append(ctx, removedLogs.Events, now); err != nil {
		return err
	}
	return a.queue.Enqueue(ctx, distinctJobs(followUp)...)
}

func distinctJobs(batch []jobs.Job) []jobs.Job {
	distinct := make([]jobs.Job, 0, len(batch))
	for _, job := range batch {
		if !slices.Contains(distinct, job) {
			distinct = append(distinct, job)
		}
	}
	return distinct
}

func FileAccountDeletion(
	ctx context.Context,
	db *pgxpool.Pool,
	filing *accountdeletion.Filing,
	request accountdeletion.Request,
) (int64, error) {
	var jobID int64
	err := postgres.RunInTransaction(ctx, db, func(ctx context.Context) error {
		var err error
		jobID, err = filing.File(ctx, request)
		return err
	})
	if err != nil {
		return 0, err
	}
	return jobID, nil
}
