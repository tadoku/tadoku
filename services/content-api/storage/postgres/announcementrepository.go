package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/content-api/domain"
)

// AnnouncementRepository implements all announcement-related domain interfaces.
type AnnouncementRepository struct {
	psql *sql.DB
	q    *Queries
}

func NewAnnouncementRepository(psql *sql.DB) *AnnouncementRepository {
	return &AnnouncementRepository{
		psql: psql,
		q:    &Queries{psql},
	}
}

func (r *AnnouncementRepository) CreateAnnouncement(ctx context.Context, a *domain.Announcement) error {
	_, err := r.q.CreateAnnouncement(ctx, CreateAnnouncementParams{
		ID:        a.ID,
		Namespace: a.Namespace,
		Title:     a.Title,
		Content:   a.Content,
		Style:     a.Style,
		Href:      NewNullString(a.Href),
		StartsAt:  a.StartsAt,
		EndsAt:    a.EndsAt,
	})
	if err != nil {
		return fmt.Errorf("could not create announcement: %w", err)
	}

	return nil
}

func (r *AnnouncementRepository) GetAnnouncementByID(ctx context.Context, id uuid.UUID, namespace string) (*domain.Announcement, error) {
	row, err := r.q.FindAnnouncementByID(ctx, FindAnnouncementByIDParams{ID: id, Namespace: namespace})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAnnouncementNotFound
		}
		return nil, fmt.Errorf("could not find announcement: %w", err)
	}

	return &domain.Announcement{
		ID:        row.ID,
		Namespace: row.Namespace,
		Title:     row.Title,
		Content:   row.Content,
		Style:     row.Style,
		Href:      NewStringFromNullString(row.Href),
		StartsAt:  row.StartsAt,
		EndsAt:    row.EndsAt,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *AnnouncementRepository) UpdateAnnouncement(ctx context.Context, a *domain.Announcement) error {
	_, err := r.q.UpdateAnnouncement(ctx, UpdateAnnouncementParams{
		ID:        a.ID,
		Namespace: a.Namespace,
		Title:     a.Title,
		Content:   a.Content,
		Style:     a.Style,
		Href:      NewNullString(a.Href),
		StartsAt:  a.StartsAt,
		EndsAt:    a.EndsAt,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAnnouncementNotFound
		}
		return fmt.Errorf("could not update announcement: %w", err)
	}

	return nil
}

func (r *AnnouncementRepository) DeleteAnnouncement(ctx context.Context, id uuid.UUID, namespace string) error {
	if err := r.q.DeleteAnnouncement(ctx, DeleteAnnouncementParams{ID: id, Namespace: namespace}); err != nil {
		return fmt.Errorf("could not delete announcement: %w", err)
	}
	return nil
}

func (r *AnnouncementRepository) ListAnnouncements(ctx context.Context, namespace string, pageSize, page int) (*domain.AnnouncementListResult, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list announcements: %w", err)
	}

	qtx := r.q.WithTx(tx)

	totalSize, err := qtx.AnnouncementsMetadata(ctx, namespace)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list announcements: %w", err)
	}

	rows, err := qtx.ListAnnouncements(ctx, ListAnnouncementsParams{
		StartFrom: int32(page * pageSize),
		PageSize:  int32(pageSize),
		Namespace: namespace,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list announcements: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not list announcements: %w", err)
	}

	result := make([]domain.Announcement, len(rows))
	for i, row := range rows {
		result[i] = domain.Announcement{
			ID:        row.ID,
			Namespace: row.Namespace,
			Title:     row.Title,
			Content:   row.Content,
			Style:     row.Style,
			Href:      NewStringFromNullString(row.Href),
			StartsAt:  row.StartsAt,
			EndsAt:    row.EndsAt,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}

	nextPageToken := ""
	if (page*pageSize)+pageSize < int(totalSize) {
		nextPageToken = fmt.Sprint(page + 1)
	}

	return &domain.AnnouncementListResult{
		Announcements: result,
		TotalSize:     int(totalSize),
		NextPageToken: nextPageToken,
	}, nil
}

func (r *AnnouncementRepository) ListActiveAnnouncements(ctx context.Context, namespace string) ([]domain.Announcement, error) {
	rows, err := r.q.ListActiveAnnouncements(ctx, namespace)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []domain.Announcement{}, nil
		}
		return nil, fmt.Errorf("could not list active announcements: %w", err)
	}

	result := make([]domain.Announcement, len(rows))
	for i, row := range rows {
		result[i] = domain.Announcement{
			ID:        row.ID,
			Namespace: row.Namespace,
			Title:     row.Title,
			Content:   row.Content,
			Style:     row.Style,
			Href:      NewStringFromNullString(row.Href),
			StartsAt:  row.StartsAt,
			EndsAt:    row.EndsAt,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}

	return result, nil
}
