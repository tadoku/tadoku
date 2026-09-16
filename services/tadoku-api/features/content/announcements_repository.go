package content

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type AnnouncementsRepository struct {
	db *pgxpool.Pool
}

func NewAnnouncementsRepository(db *pgxpool.Pool) *AnnouncementsRepository {
	return &AnnouncementsRepository{
		db: db,
	}
}

func (r *AnnouncementsRepository) DeleteAnnouncement(ctx context.Context, namespace string, id uuid.UUID, deletedAt time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).DeleteAnnouncement(ctx, queries.DeleteAnnouncementParams{
		Namespace: namespace,
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		DeletedAt: pgtype.Timestamp{Time: deletedAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}

	return nil
}

func (r *AnnouncementsRepository) FindAnnouncementByID(ctx context.Context, namespace string, id uuid.UUID) (*Announcement, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindAnnouncementByID(ctx, queries.FindAnnouncementByIDParams{
		Namespace: namespace,
		ID:        pgtype.UUID{Bytes: id, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAnnouncementNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find announcement by ID: %w", err)
	}

	var href *string
	if row.Href.Valid {
		href = &row.Href.String
	}

	return &Announcement{
		ID:        uuid.UUID(row.ID.Bytes),
		Namespace: row.Namespace,
		Title:     row.Title,
		Content:   row.Content,
		Style:     row.Style,
		Href:      href,
		StartsAt:  row.StartsAt.Time,
		EndsAt:    row.EndsAt.Time,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

func (r *AnnouncementsRepository) CountAnnouncements(ctx context.Context, namespace string) (int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return 0, err
	}

	total, err := queries.New(executor).CountAnnouncements(ctx, namespace)
	if err != nil {
		return 0, fmt.Errorf("count announcements: %w", err)
	}
	return int(total), nil
}

func (r *AnnouncementsRepository) ListAnnouncements(ctx context.Context, namespace string, limit, offset int32) ([]Announcement, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListAnnouncements(ctx, queries.ListAnnouncementsParams{
		Namespace:   namespace,
		ResultLimit: limit,
		StartFrom:   offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}

	result := make([]Announcement, 0, len(rows))
	for _, row := range rows {
		var href *string
		if row.Href.Valid {
			href = &row.Href.String
		}

		result = append(result, Announcement{
			ID:        uuid.UUID(row.ID.Bytes),
			Namespace: row.Namespace,
			Title:     row.Title,
			Content:   row.Content,
			Style:     row.Style,
			Href:      href,
			StartsAt:  row.StartsAt.Time,
			EndsAt:    row.EndsAt.Time,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		})
	}

	return result, nil
}

func (r *AnnouncementsRepository) ListActiveAnnouncements(ctx context.Context, namespace string, cutoff time.Time, limit int32) ([]Announcement, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListActiveAnnouncements(ctx, queries.ListActiveAnnouncementsParams{
		Namespace:   namespace,
		Cutoff:      pgtype.Timestamp{Time: cutoff, Valid: true},
		ResultLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list active announcements: %w", err)
	}

	result := make([]Announcement, 0, len(rows))
	for _, row := range rows {
		var href *string
		if row.Href.Valid {
			href = &row.Href.String
		}

		result = append(result, Announcement{
			ID:        uuid.UUID(row.ID.Bytes),
			Namespace: row.Namespace,
			Title:     row.Title,
			Content:   row.Content,
			Style:     row.Style,
			Href:      href,
			StartsAt:  row.StartsAt.Time,
			EndsAt:    row.EndsAt.Time,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		})
	}

	return result, nil
}
