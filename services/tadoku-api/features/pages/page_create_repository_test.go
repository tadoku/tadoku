package pages_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestPagesRepositoryCreatePageIsAtomic(t *testing.T) {
	t.Parallel()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	item := &pages.Page{
		ID:          uuid.New(),
		Namespace:   "main",
		Slug:        "created-page",
		Title:       "Created",
		HTML:        "<p>Created</p>",
		PublishedAt: &instant,
		CreatedAt:   &instant,
		UpdatedAt:   &instant,
	}
	repository := pages.NewPagesRepository(db.Pool)
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		if err := repository.CreatePage(ctx, item); err != nil {
			return err
		}
		got, err := repository.FindPageByID(ctx, item.Namespace, item.ID)
		if err != nil {
			return err
		}
		if got.Title != item.Title || got.HTML != item.HTML {
			t.Errorf("transaction readback=%+v", got)
		}
		if got.PublishedAt == nil || !got.PublishedAt.Equal(instant) {
			t.Errorf("published_at=%v, want %v", got.PublishedAt, instant)
		}
		if got.CreatedAt == nil || !got.CreatedAt.Equal(instant) {
			t.Errorf("created_at=%v, want %v", got.CreatedAt, instant)
		}
		if got.UpdatedAt == nil || !got.UpdatedAt.Equal(instant) {
			t.Errorf("updated_at=%v, want %v", got.UpdatedAt, instant)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	var pagesCount, versions int
	if err := db.Pool.QueryRow(t.Context(), "select (select count(*) from pages), (select count(*) from pages_content)").Scan(&pagesCount, &versions); err != nil {
		t.Fatal(err)
	}
	if pagesCount != 1 || versions != 1 {
		t.Errorf("created %d pages and %d versions, want one each", pagesCount, versions)
	}

	duplicate := *item
	duplicate.ID = uuid.New()
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.CreatePage(ctx, &duplicate)
	})
	if !errors.Is(err, pages.ErrPageAlreadyExists) {
		t.Errorf("duplicate slug error=%v, want conflict", err)
	}
}

func TestPagesRepositoryCreatePageRollsBackWhenContentFails(t *testing.T) {
	t.Parallel()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Pool.Exec(t.Context(), "alter table pages_content add constraint reject_page_title check (title <> 'Rejected')"); err != nil {
		t.Fatal(err)
	}

	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	item := &pages.Page{
		ID:        uuid.New(),
		Namespace: "main",
		Slug:      "failed-page",
		Title:     "Rejected",
		HTML:      "<p>Rejected</p>",
		CreatedAt: &instant,
		UpdatedAt: &instant,
	}
	repository := pages.NewPagesRepository(db.Pool)
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.CreatePage(ctx, item)
	})
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != pgerrcode.CheckViolation {
		t.Fatalf("error=%v, want rejected content version", err)
	}

	var pagesCount, versions int
	if err := db.Pool.QueryRow(t.Context(), "select (select count(*) from pages), (select count(*) from pages_content)").Scan(&pagesCount, &versions); err != nil {
		t.Fatal(err)
	}
	if pagesCount != 0 || versions != 0 {
		t.Errorf("failed create left %d pages and %d versions", pagesCount, versions)
	}
}
