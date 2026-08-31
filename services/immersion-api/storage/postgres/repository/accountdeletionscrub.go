package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) ScrubAccount(ctx context.Context, req *domain.AccountDeletionScrubRequest) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.q.WithTx(tx)
	target, err := qtx.LockAccountDeletionTarget(ctx, req.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrAccountDeletionNotLocked
	}
	if err != nil {
		return fmt.Errorf("could not lock deletion target: %w", err)
	}
	if target.DeletedAt.Valid {
		return tx.Commit()
	}
	if !target.DeletionLockedAt.Valid {
		return domain.ErrAccountDeletionNotLocked
	}

	hasRunningContest, err := qtx.HasRunningOwnedContest(ctx, postgres.HasRunningOwnedContestParams{
		UserID:    req.UserID,
		DeletedAt: req.DeletedAt(),
	})
	if err != nil {
		return fmt.Errorf("could not classify owned contests: %w", err)
	}
	if hasRunningContest {
		return domain.ErrRunningContestOwned
	}

	contestIDs, err := qtx.ListNonHistoricalContestIDsForAccount(ctx, postgres.ListNonHistoricalContestIDsForAccountParams{
		UserID:    req.UserID,
		DeletedAt: req.DeletedAt(),
	})
	if err != nil {
		return fmt.Errorf("could not list non-historical contests: %w", err)
	}
	years, err := qtx.ListOfficialLeaderboardYearsForAccount(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("could not list official leaderboard years: %w", err)
	}

	params := postgres.DeleteNonHistoricalRegistrationsForAccountParams{UserID: req.UserID, DeletedAt: req.DeletedAt()}
	if err := qtx.DeleteNonHistoricalRegistrationsForAccount(ctx, params); err != nil {
		return fmt.Errorf("could not delete non-historical registrations: %w", err)
	}
	if err := qtx.DetachNonHistoricalLogsForAccount(ctx, postgres.DetachNonHistoricalLogsForAccountParams{UserID: req.UserID, DeletedAt: req.DeletedAt()}); err != nil {
		return fmt.Errorf("could not detach non-historical logs: %w", err)
	}
	if err := qtx.CancelFutureOwnedContestsForAccount(ctx, postgres.CancelFutureOwnedContestsForAccountParams{
		CanceledAt: req.DeletedAt(),
		UserID:     req.UserID,
		Today:      req.DeletedAt(),
	}); err != nil {
		return fmt.Errorf("could not cancel future contests: %w", err)
	}
	if err := qtx.AnonymizeOwnedContestsForAccount(ctx, req.UserID); err != nil {
		return fmt.Errorf("could not anonymize contest ownership: %w", err)
	}
	if err := qtx.FreezeHistoricalLogsForAccount(ctx, postgres.FreezeHistoricalLogsForAccountParams{
		UserID:    req.UserID,
		DeletedAt: sql.NullTime{Time: req.DeletedAt(), Valid: true},
	}); err != nil {
		return fmt.Errorf("could not freeze historical logs: %w", err)
	}
	if err := qtx.DeleteTagsForAccount(ctx, req.UserID); err != nil {
		return fmt.Errorf("could not delete log tags: %w", err)
	}
	if err := qtx.DeleteNonHistoricalLogsForAccount(ctx, postgres.DeleteNonHistoricalLogsForAccountParams{UserID: req.UserID, DeletedAt: req.DeletedAt()}); err != nil {
		return fmt.Errorf("could not delete non-historical logs: %w", err)
	}
	if err := qtx.MarkAccountDeleted(ctx, postgres.MarkAccountDeletedParams{UserID: req.UserID, DeletedAt: req.DeletedAt()}); err != nil {
		return fmt.Errorf("could not mark account deleted: %w", err)
	}

	for _, contestID := range contestIDs {
		if err := insertAccountDeletionOutboxEvent(ctx, qtx, "remove_contest_score", req.UserID, &contestID, nil); err != nil {
			return err
		}
	}
	for _, year := range years {
		year := year
		if err := insertAccountDeletionOutboxEvent(ctx, qtx, "remove_official_scores", req.UserID, nil, &year); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit account scrub: %w", err)
	}
	return nil
}

func insertAccountDeletionOutboxEvent(ctx context.Context, qtx *postgres.Queries, eventType string, userID uuid.UUID, contestID *uuid.UUID, year *int16) error {
	contest := uuid.NullUUID{}
	if contestID != nil {
		contest = uuid.NullUUID{UUID: *contestID, Valid: true}
	}
	y := sql.NullInt16{}
	if year != nil {
		y = sql.NullInt16{Int16: *year, Valid: true}
	}
	if err := qtx.InsertLeaderboardOutboxEvent(ctx, postgres.InsertLeaderboardOutboxEventParams{
		EventType: eventType,
		UserID:    userID,
		ContestID: contest,
		Year:      y,
	}); err != nil {
		return fmt.Errorf("could not insert %s outbox event: %w", eventType, err)
	}
	return nil
}
