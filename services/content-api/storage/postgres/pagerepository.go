package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/tadoku/tadoku/services/content-api/domain/pagecommand"
	"github.com/tadoku/tadoku/services/content-api/domain/pagequery"
)

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

func (r *PageRepository) CreatePage(ctx context.Context, req *pagecommand.PageCreateRequest) (*pagecommand.PageCreateResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	pageContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	_, err = qtx.CreatePage(ctx, CreatePageParams{
		req.ID,
		req.Slug,
		pageContentID,
		NewNullTime(req.PublishedAt),
	})
	if err != nil {
		_ = tx.Rollback()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, pagecommand.ErrPageAlreadyExists
		}

		return nil, fmt.Errorf("could not create page: %w", err)
	}

	_, err = qtx.CreatePageContent(ctx, CreatePageContentParams{
		ID:     pageContentID,
		PageID: req.ID,
		Title:  req.Title,
		Html:   req.Html,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	page, err := qtx.FindPageBySlug(ctx, req.Slug)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	return &pagecommand.PageCreateResponse{
		ID:          page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Html:        page.Html,
		PublishedAt: NewTimeFromNullTime(page.PublishedAt),
	}, nil
}

func (r *PageRepository) UpdatePage(ctx context.Context, id uuid.UUID, req *pagecommand.PageUpdateRequest) (*pagecommand.PageUpdateResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	pageContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	_, err = qtx.UpdatePage(ctx, UpdatePageParams{
		ID:               id,
		Slug:             req.Slug,
		CurrentContentID: pageContentID,
		PublishedAt:      NewNullTime(req.PublishedAt),
	})
	if err != nil {
		_ = tx.Rollback()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, pagecommand.ErrPageAlreadyExists
		}

		return nil, fmt.Errorf("could not create page: %w", err)
	}

	_, err = qtx.CreatePageContent(ctx, CreatePageContentParams{
		ID:     pageContentID,
		PageID: id,
		Title:  req.Title,
		Html:   req.Html,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	page, err := qtx.FindPageBySlug(ctx, req.Slug)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	return &pagecommand.PageUpdateResponse{
		ID:          page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Html:        page.Html,
		PublishedAt: NewTimeFromNullTime(page.PublishedAt),
	}, nil
}

func (r *PageRepository) FindBySlug(ctx context.Context, slug string) (*pagequery.PageFindResponse, error) {
	page, err := r.q.FindPageBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pagequery.ErrPageNotFound
		}

		return nil, fmt.Errorf("could not find page: %w", err)
	}

	return &pagequery.PageFindResponse{
		ID:          page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Html:        page.Html,
		PublishedAt: NewTimeFromNullTime(page.PublishedAt),
	}, nil
}

func (r *PageRepository) ListPages(ctx context.Context, req *pagequery.PageListRequest) (*pagequery.PageListResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list pages: %w", err)
	}

	qtx := r.q.WithTx(tx)

	meta, err := qtx.PagesMetadata(ctx, req.IncludeDrafts)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not lists pages: %w", err)
	}

	pages, err := qtx.ListPages(ctx, ListPagesParams{
		StartFrom:     int32(req.Page * req.PageSize),
		PageSize:      int32(req.PageSize),
		IncludeDrafts: req.IncludeDrafts,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list pages: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not list pages: %w", err)
	}

	res := make([]pagequery.PageListEntry, len(pages))
	for i, page := range pages {
		res[i] = pagequery.PageListEntry{
			ID:          page.ID,
			Slug:        page.Slug,
			Title:       page.Title,
			PublishedAt: NewTimeFromNullTime(page.PublishedAt),
			CreatedAt:   page.CreatedAt,
			UpdatedAt:   page.UpdatedAt,
		}
	}

	nextPageToken := ""
	if (req.Page*req.PageSize)+req.PageSize < int(meta.TotalSize) {
		nextPageToken = fmt.Sprint(req.Page + 1)
	}

	return &pagequery.PageListResponse{
		Pages:         res,
		TotalSize:     int(meta.TotalSize),
		NextPageToken: nextPageToken,
	}, nil
}
