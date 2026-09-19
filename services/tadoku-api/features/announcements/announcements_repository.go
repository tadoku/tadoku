package announcements

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/announcements"
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
		DeletedAt: postgres.Timestamp(deletedAt),
	})
	if err != nil {
		return fmt.Errorf("delete announcement: %w", err)
	}

	return nil
}

func (r *AnnouncementsRepository) CreateAnnouncement(ctx context.Context, item *Announcement) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var href pgtype.Text
	if item.Href != nil {
		href = pgtype.Text{String: *item.Href, Valid: true}
	}
	err = queries.New(executor).CreateAnnouncement(ctx, queries.CreateAnnouncementParams{
		ID:        pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace: item.Namespace,
		Title:     item.Title,
		Content:   item.Content,
		Style:     item.Style,
		Href:      href,
		StartsAt:  postgres.Timestamp(item.StartsAt),
		EndsAt:    postgres.Timestamp(item.EndsAt),
		CreatedAt: postgres.Timestamp(item.CreatedAt),
		UpdatedAt: postgres.Timestamp(item.UpdatedAt),
	})
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrAnnouncementAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create announcement: %w", err)
	}
	return nil
}

func (r *AnnouncementsRepository) UpdateAnnouncement(ctx context.Context, item *Announcement) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var href pgtype.Text
	if item.Href != nil {
		href = pgtype.Text{String: *item.Href, Valid: true}
	}
	_, err = queries.New(executor).UpdateAnnouncement(ctx, queries.UpdateAnnouncementParams{
		ID:        pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace: item.Namespace,
		Title:     item.Title,
		Content:   item.Content,
		Style:     item.Style,
		Href:      href,
		StartsAt:  postgres.Timestamp(item.StartsAt),
		EndsAt:    postgres.Timestamp(item.EndsAt),
		UpdatedAt: postgres.Timestamp(item.UpdatedAt),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrAnnouncementNotFound
	}
	if err != nil {
		return fmt.Errorf("update announcement: %w", err)
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

func (r *AnnouncementsRepository) ListAnnouncements(ctx context.Context, namespace string, limit int32, offset int64) ([]Announcement, int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	rows, err := queries.New(executor).ListAnnouncements(ctx, queries.ListAnnouncementsParams{
		Namespace:   namespace,
		ResultLimit: limit,
		StartFrom:   offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list announcements: %w", err)
	}

	result := make([]Announcement, 0, len(rows))
	var total int
	for _, row := range rows {
		total = int(row.TotalSize)
		if !row.ID.Valid {
			continue
		}

		var href *string
		if row.Href.Valid {
			href = &row.Href.String
		}

		result = append(result, Announcement{
			ID:        uuid.UUID(row.ID.Bytes),
			Namespace: row.Namespace.String,
			Title:     row.Title.String,
			Content:   row.Content.String,
			Style:     row.Style.String,
			Href:      href,
			StartsAt:  row.StartsAt.Time,
			EndsAt:    row.EndsAt.Time,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		})
	}

	return result, total, nil
}

func (r *AnnouncementsRepository) ListActiveAnnouncements(ctx context.Context, namespace string, cutoff time.Time, limit int32) ([]Announcement, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListActiveAnnouncements(ctx, queries.ListActiveAnnouncementsParams{
		Namespace:   namespace,
		Cutoff:      postgres.Timestamp(cutoff),
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
