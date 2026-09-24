package pages

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
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type PagesRepository struct {
	db *pgxpool.Pool
}

func NewPagesRepository(db *pgxpool.Pool) *PagesRepository {
	return &PagesRepository{db: db}
}

func (r *PagesRepository) CreatePage(ctx context.Context, item *Page, contentID uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var publishedAt pgtype.Timestamp
	if item.PublishedAt != nil {
		publishedAt = postgres.Timestamp(*item.PublishedAt)
	}
	err = queries.New(executor).CreatePage(ctx, queries.CreatePageParams{
		ID:               pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace:        item.Namespace,
		Slug:             item.Slug,
		CurrentContentID: pgtype.UUID{Bytes: contentID, Valid: true},
		PublishedAt:      publishedAt,
		CreatedAt:        postgres.Timestamp(*item.CreatedAt),
		UpdatedAt:        postgres.Timestamp(*item.UpdatedAt),
	})
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrPageAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create page: %w", err)
	}

	return nil
}

func (r *PagesRepository) CreatePageContent(ctx context.Context, pageID, contentID uuid.UUID, title, html string, createdAt time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).CreatePageContent(ctx, queries.CreatePageContentParams{
		ID:        pgtype.UUID{Bytes: contentID, Valid: true},
		PageID:    pgtype.UUID{Bytes: pageID, Valid: true},
		Title:     title,
		Html:      html,
		CreatedAt: postgres.Timestamp(createdAt),
	})
	if err != nil {
		return fmt.Errorf("create page content: %w", err)
	}

	return nil
}

func (r *PagesRepository) DeletePage(ctx context.Context, namespace string, id uuid.UUID, deletedAt time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).DeletePage(ctx, queries.DeletePageParams{
		Namespace: namespace,
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		DeletedAt: postgres.Timestamp(deletedAt),
	})
	if err != nil {
		return fmt.Errorf("delete page: %w", err)
	}

	return nil
}

func (r *PagesRepository) FindPageBySlug(ctx context.Context, namespace, slug string) (*Page, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindPageBySlug(ctx, queries.FindPageBySlugParams{
		Namespace: namespace,
		Slug:      slug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find page by slug: %w", err)
	}
	return pageFromFindRow(row), nil
}

func (r *PagesRepository) FindPageByID(ctx context.Context, namespace string, id uuid.UUID) (*Page, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindPageByID(ctx, queries.FindPageByIDParams{
		Namespace: namespace,
		ID:        pgtype.UUID{Bytes: id, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find page by ID: %w", err)
	}
	return pageFromFindRow(queries.FindPageBySlugRow(row)), nil
}

func pageFromFindRow(row queries.FindPageBySlugRow) *Page {
	var publishedAt *time.Time
	if row.PublishedAt.Valid {
		publishedAt = &row.PublishedAt.Time
	}
	return &Page{
		ID:          uuid.UUID(row.ID.Bytes),
		Namespace:   row.Namespace,
		Slug:        row.Slug,
		Title:       row.Title,
		HTML:        row.Html,
		PublishedAt: publishedAt,
		CreatedAt:   &row.CreatedAt.Time,
		UpdatedAt:   &row.UpdatedAt.Time,
	}
}

func (r *PagesRepository) ListPages(ctx context.Context, namespace string, includeDrafts bool, limit int32, offset int64) ([]Page, int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	rows, err := queries.New(executor).ListPages(ctx, queries.ListPagesParams{
		Namespace:     namespace,
		IncludeDrafts: includeDrafts,
		ResultLimit:   limit,
		StartFrom:     offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list pages: %w", err)
	}

	result := make([]Page, 0, len(rows))
	var total int
	for _, row := range rows {
		total = int(row.TotalSize)
		if !row.ID.Valid {
			continue
		}

		var publishedAt *time.Time
		if row.PublishedAt.Valid {
			publishedAt = &row.PublishedAt.Time
		}
		result = append(result, Page{
			ID:          uuid.UUID(row.ID.Bytes),
			Namespace:   row.Namespace.String,
			Slug:        row.Slug.String,
			Title:       row.Title.String,
			HTML:        row.Html.String,
			PublishedAt: publishedAt,
			CreatedAt:   &row.CreatedAt.Time,
			UpdatedAt:   &row.UpdatedAt.Time,
		})
	}

	return result, total, nil
}

func (r *PagesRepository) UpdatePage(ctx context.Context, item *Page, contentID *uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var publishedAt pgtype.Timestamp
	if item.PublishedAt != nil {
		publishedAt = postgres.Timestamp(*item.PublishedAt)
	}
	_, err = queries.New(executor).UpdatePage(ctx, queries.UpdatePageParams{
		ID:               pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace:        item.Namespace,
		Slug:             item.Slug,
		CurrentContentID: postgres.NullableUUID(contentID),
		PublishedAt:      publishedAt,
		UpdatedAt:        postgres.Timestamp(*item.UpdatedAt),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPageNotFound
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrPageAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("update page: %w", err)
	}

	return nil
}

func (r *PagesRepository) GetPageVersion(ctx context.Context, namespace string, pageID, contentID uuid.UUID) (*PageVersion, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).GetPageVersion(ctx, queries.GetPageVersionParams{
		Namespace: namespace,
		PageID:    pgtype.UUID{Bytes: pageID, Valid: true},
		ContentID: pgtype.UUID{Bytes: contentID, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get page version: %w", err)
	}

	return &PageVersion{
		ID:        uuid.UUID(row.ID.Bytes),
		Version:   int(row.Version),
		Title:     row.Title,
		HTML:      row.Html,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (r *PagesRepository) ListPageVersions(ctx context.Context, namespace string, id uuid.UUID) ([]PageVersion, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListPageVersions(ctx, queries.ListPageVersionsParams{
		Namespace: namespace,
		PageID:    pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list page versions: %w", err)
	}

	versions := make([]PageVersion, 0, len(rows))
	for i, row := range rows {
		versions = append(versions, PageVersion{
			ID:        uuid.UUID(row.ID.Bytes),
			Version:   i + 1,
			Title:     row.Title,
			CreatedAt: row.CreatedAt.Time,
		})
	}

	return versions, nil
}
