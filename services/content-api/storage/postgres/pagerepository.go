package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/tadoku/tadoku/services/content-api/domain"
)

// PageRepository implements all page-related domain interfaces:
// - domain.PageCreateRepository
// - domain.PageUpdateRepository
// - domain.PageFindRepository
// - domain.PageListRepository
type PageRepository struct {
	psql *sql.DB
	q    *Queries
}

func NewPageRepository(psql *sql.DB) *PageRepository {
	return &PageRepository{
		psql: psql,
		q:    &Queries{psql},
	}
}

// CreatePage implements domain.PageCreateRepository
func (r *PageRepository) CreatePage(ctx context.Context, page *domain.Page) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}

	pageContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	_, err = qtx.CreatePage(ctx, CreatePageParams{
		ID:               page.ID,
		Namespace:        page.Namespace,
		Slug:             page.Slug,
		CurrentContentID: pageContentID,
		PublishedAt:      NewNullTime(page.PublishedAt),
	})
	if err != nil {
		_ = tx.Rollback()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrPageAlreadyExists
		}

		return fmt.Errorf("could not create page: %w", err)
	}

	_, err = qtx.CreatePageContent(ctx, CreatePageContentParams{
		ID:     pageContentID,
		PageID: page.ID,
		Title:  page.Title,
		Html:   page.HTML,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not create page: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}

	return nil
}

// GetPageByID implements domain.PageUpdateRepository
func (r *PageRepository) GetPageByID(ctx context.Context, id uuid.UUID) (*domain.Page, error) {
	page, err := r.q.FindPageByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPageNotFound
		}
		return nil, fmt.Errorf("could not find page: %w", err)
	}

	return &domain.Page{
		ID:          page.ID,
		Namespace:   page.Namespace,
		Slug:        page.Slug,
		Title:       page.Title,
		HTML:        page.Html,
		PublishedAt: NewTimeFromNullTime(page.PublishedAt),
		CreatedAt:   page.CreatedAt,
		UpdatedAt:   page.UpdatedAt,
	}, nil
}

// UpdatePage implements domain.PageUpdateRepository
func (r *PageRepository) UpdatePage(ctx context.Context, page *domain.Page) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not update page: %w", err)
	}

	pageContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	_, err = qtx.UpdatePage(ctx, UpdatePageParams{
		ID:               page.ID,
		Slug:             page.Slug,
		CurrentContentID: pageContentID,
		PublishedAt:      NewNullTime(page.PublishedAt),
	})
	if err != nil {
		_ = tx.Rollback()

		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrPageNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrPageAlreadyExists
		}

		return fmt.Errorf("could not update page: %w", err)
	}

	_, err = qtx.CreatePageContent(ctx, CreatePageContentParams{
		ID:     pageContentID,
		PageID: page.ID,
		Title:  page.Title,
		Html:   page.HTML,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not update page: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not update page: %w", err)
	}

	return nil
}

// DeletePage implements domain.PageDeleteRepository
func (r *PageRepository) DeletePage(ctx context.Context, id uuid.UUID) error {
	if err := r.q.DeletePage(ctx, id); err != nil {
		return fmt.Errorf("could not delete page: %w", err)
	}
	return nil
}

// FindPageBySlug implements domain.PageFindRepository
func (r *PageRepository) FindPageBySlug(ctx context.Context, namespace, slug string) (*domain.Page, error) {
	page, err := r.q.FindPageBySlug(ctx, FindPageBySlugParams{
		Namespace: namespace,
		Slug:      slug,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPageNotFound
		}

		return nil, fmt.Errorf("could not find page: %w", err)
	}

	return &domain.Page{
		ID:          page.ID,
		Namespace:   page.Namespace,
		Slug:        page.Slug,
		Title:       page.Title,
		HTML:        page.Html,
		PublishedAt: NewTimeFromNullTime(page.PublishedAt),
		CreatedAt:   page.CreatedAt,
		UpdatedAt:   page.UpdatedAt,
	}, nil
}

// ListPages implements domain.PageListRepository
func (r *PageRepository) ListPages(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*domain.PageListResult, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list pages: %w", err)
	}

	qtx := r.q.WithTx(tx)

	meta, err := qtx.PagesMetadata(ctx, PagesMetadataParams{
		IncludeDrafts: includeDrafts,
		Namespace:     namespace,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list pages: %w", err)
	}

	pages, err := qtx.ListPages(ctx, ListPagesParams{
		StartFrom:     int32(page * pageSize),
		PageSize:      int32(pageSize),
		IncludeDrafts: includeDrafts,
		Namespace:     namespace,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list pages: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not list pages: %w", err)
	}

	result := make([]domain.Page, len(pages))
	for i, p := range pages {
		result[i] = domain.Page{
			ID:          p.ID,
			Namespace:   p.Namespace,
			Slug:        p.Slug,
			Title:       p.Title,
			HTML:        p.Html,
			PublishedAt: NewTimeFromNullTime(p.PublishedAt),
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	nextPageToken := ""
	if (page*pageSize)+pageSize < int(meta.TotalSize) {
		nextPageToken = fmt.Sprint(page + 1)
	}

	return &domain.PageListResult{
		Pages:         result,
		TotalSize:     int(meta.TotalSize),
		NextPageToken: nextPageToken,
	}, nil
}
