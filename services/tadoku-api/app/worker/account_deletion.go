package worker

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

// DeleteAccount locks the account row and then the user row, the order of every transaction that locks both, so it
// waits for a login, recovery, settings write or filing that holds the account row instead of deadlocking with it.
func (a *Application) DeleteAccount(ctx context.Context, job jobs.DeleteAccountV1) error {
	now := timex.Now()
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.identityDeletion.LockAccountForDeletion(ctx, job.UserID); err != nil {
			return err
		}
		tombstoned, err := a.profile.Tombstone(ctx, job.UserID, now)
		if err != nil || !tombstoned {
			return err
		}
		return a.removeAccountData(ctx, job.UserID, now)
	})
	if errors.Is(err, profile.ErrDeletionNotRequested) || errors.Is(err, profile.ErrLocalUserNotFound) {
		return rejected(err)
	}
	return err
}

func (a *Application) removeAccountData(ctx context.Context, userID uuid.UUID, now time.Time) error {
	removedLogs, err := a.logs.RemoveUser(ctx, userID, now)
	if err != nil {
		return err
	}
	removedRegistrations, err := a.contests.RemoveUser(ctx, userID, now)
	if err != nil {
		return err
	}
	if err := a.relationships.RevokeUser(ctx, userID); err != nil {
		return err
	}
	notice, hasNotice, err := a.identityDeletion.DeleteAccount(ctx, userID, now)
	if err != nil {
		return err
	}

	followUp := slices.Concat(removedLogs.Jobs, removedRegistrations.Jobs)
	if hasNotice {
		written, err := a.email.Write(ctx, notice)
		if err != nil {
			return err
		}
		followUp = append(followUp, written.Jobs...)
	}

	err = a.audit.Record(ctx, audit.Event{
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
	if err := a.contestEvents.Append(ctx, removedLogs.Events, now); err != nil {
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
