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

func (r *AnnouncementsRepository) CreateAnnouncement(ctx context.Context, item *Announcement) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).CreateAnnouncement(ctx, queries.CreateAnnouncementParams{
		ID:        pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace: item.Namespace,
		Title:     item.Title,
		Content:   item.Content,
		Style:     item.Style,
		Href:      hrefToText(item.Href),
		StartsAt:  pgtype.Timestamp{Time: item.StartsAt, Valid: true},
		EndsAt:    pgtype.Timestamp{Time: item.EndsAt, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: item.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: item.UpdatedAt, Valid: true},
	})
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

	_, err = queries.New(executor).UpdateAnnouncement(ctx, queries.UpdateAnnouncementParams{
		ID:        pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace: item.Namespace,
		Title:     item.Title,
		Content:   item.Content,
		Style:     item.Style,
		Href:      hrefToText(item.Href),
		StartsAt:  pgtype.Timestamp{Time: item.StartsAt, Valid: true},
		EndsAt:    pgtype.Timestamp{Time: item.EndsAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: item.UpdatedAt, Valid: true},
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

	item := announcementFrom(row.ID, row.Namespace, row.Title, row.Content, row.Style, row.Href, row.StartsAt, row.EndsAt, row.CreatedAt, row.UpdatedAt)
	return &item, nil
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
		result = append(result, announcementFrom(row.ID, row.Namespace, row.Title, row.Content, row.Style, row.Href, row.StartsAt, row.EndsAt, row.CreatedAt, row.UpdatedAt))
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
		result = append(result, announcementFrom(row.ID, row.Namespace, row.Title, row.Content, row.Style, row.Href, row.StartsAt, row.EndsAt, row.CreatedAt, row.UpdatedAt))
	}

	return result, nil
}

func hrefFromText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	href := value.String
	return &href
}

func hrefToText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func announcementFrom(id pgtype.UUID, namespace, title, content, style string, href pgtype.Text, startsAt, endsAt, createdAt, updatedAt pgtype.Timestamp) Announcement {
	return Announcement{
		ID:        uuid.UUID(id.Bytes),
		Namespace: namespace,
		Title:     title,
		Content:   content,
		Style:     style,
		Href:      hrefFromText(href),
		StartsAt:  startsAt.Time,
		EndsAt:    endsAt.Time,
		CreatedAt: createdAt.Time,
		UpdatedAt: updatedAt.Time,
	}
}
