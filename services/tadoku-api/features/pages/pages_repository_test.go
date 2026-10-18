package pages_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestPagesRepositoryReadsCurrentLiveContent(t *testing.T) {
	t.Parallel()
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db, err := testpostgres.New(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	_, err = db.Pool.Exec(tenantCtx, `
		insert into pages (
			tenant, id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at
		)
		values
			('tadoku/prod', '10000000-0000-4000-8000-000000000001', 'main',
			 'first-page', '20000000-0000-4000-8000-000000000002', '2026-09-12 12:00:01',
			 '2026-09-12 11:00:00', '2026-09-12 11:30:00', null),
			('tadoku/prod', '10000000-0000-4000-8000-000000000002', 'main',
			 'draft-page', '20000000-0000-4000-8000-000000000003', null,
			 '2026-09-12 11:00:00', '2026-09-12 11:00:00', null),
			('tadoku/prod', '10000000-0000-4000-8000-000000000003', 'other',
			 'other-page', '20000000-0000-4000-8000-000000000004', '2026-09-12 10:00:00',
			 '2026-09-12 11:00:00', '2026-09-12 11:00:00', null),
			('tadoku/prod', '10000000-0000-4000-8000-000000000004', 'main',
			 'deleted-page', '20000000-0000-4000-8000-000000000005', '2026-09-12 10:00:00',
			 '2026-09-12 12:00:00', '2026-09-12 12:00:00', '2026-09-12 12:30:00');
		insert into pages_content (tenant, id, page_id, title, html, created_at)
		values
			('tadoku/prod', '20000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000001',
			 'Old title', '<p>Old</p>', '2026-09-12 10:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000001',
			 'Current title', '<p>Current</p>', '2026-09-12 11:30:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000003', '10000000-0000-4000-8000-000000000002',
			 'Draft', '<p>Draft</p>', '2026-09-12 11:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000004', '10000000-0000-4000-8000-000000000003',
			 'Other', '<p>Other</p>', '2026-09-12 11:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000005', '10000000-0000-4000-8000-000000000004',
			 'Deleted', '<p>Deleted</p>', '2026-09-12 12:00:00')`)
	if err != nil {
		t.Fatal(err)
	}

	repository := pages.NewPagesRepository(db.Pool)
	page, err := repository.FindPageBySlug(tenantCtx, "main", "first-page")
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "Current title" || page.HTML != "<p>Current</p>" {
		t.Errorf("current page=%+v", page)
	}

	for _, lookup := range []struct{ namespace, slug string }{
		{namespace: "other", slug: "first-page"},
		{namespace: "main", slug: "deleted-page"},
	} {
		if _, err := repository.FindPageBySlug(tenantCtx, lookup.namespace, lookup.slug); !errors.Is(err, pages.ErrPageNotFound) {
			t.Errorf("find %s/%s error=%v, want not found", lookup.namespace, lookup.slug, err)
		}
	}

	cutoff := time.Date(2026, 9, 12, 12, 0, 1, 0, time.UTC)
	for _, test := range []struct {
		name          string
		includeDrafts bool
		cutoff        time.Time
		limit         int32
		offset        int64
		wantIDs       []uuid.UUID
		wantTotal     int
	}{
		{name: "all live rows with stable ties", includeDrafts: true, cutoff: cutoff.Add(-time.Second), limit: 10, wantIDs: []uuid.UUID{
			uuid.MustParse("10000000-0000-4000-8000-000000000002"),
			uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		}, wantTotal: 2},
		{name: "published by cutoff", cutoff: cutoff, limit: 10, wantIDs: []uuid.UUID{
			uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		}, wantTotal: 1},
		{name: "before cutoff", cutoff: cutoff.Add(-time.Second), limit: 10, wantIDs: []uuid.UUID{}, wantTotal: 0},
		{name: "empty page preserves total", includeDrafts: true, cutoff: cutoff, limit: 10, offset: 100, wantIDs: []uuid.UUID{}, wantTotal: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			items, total, err := repository.ListPages(tenantCtx, "main", test.includeDrafts, test.cutoff, test.limit, test.offset)
			if err != nil {
				t.Fatal(err)
			}

			ids := make([]uuid.UUID, len(items))
			for i := range items {
				ids[i] = items[i].ID
			}
			if !reflect.DeepEqual(ids, test.wantIDs) || total != test.wantTotal {
				t.Errorf("ids=%v total=%d, want ids=%v total=%d", ids, total, test.wantIDs, test.wantTotal)
			}
		})
	}
}

func TestPagesRepositoryCreatePageIsAtomic(t *testing.T) {
	t.Parallel()
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db, err := testpostgres.New(tenantCtx)
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
	if err := postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
		contentID := uuid.New()
		if err := repository.CreatePage(ctx, item, contentID); err != nil {
			return err
		}
		if err := repository.CreatePageContent(ctx, item.ID, contentID, item.Title, item.HTML, *item.CreatedAt); err != nil {
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
	if err := db.Pool.QueryRow(tenantCtx, "select (select count(*) from pages), (select count(*) from pages_content)").Scan(&pagesCount, &versions); err != nil {
		t.Fatal(err)
	}
	if pagesCount != 1 || versions != 1 {
		t.Errorf("created %d pages and %d versions, want one each", pagesCount, versions)
	}

	duplicate := *item
	duplicate.ID = uuid.New()
	err = postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
		return repository.CreatePage(ctx, &duplicate, uuid.New())
	})
	if !errors.Is(err, pages.ErrPageAlreadyExists) {
		t.Errorf("duplicate slug error=%v, want conflict", err)
	}
}

func TestPagesRepositoryCreatePageRollsBackWhenContentFails(t *testing.T) {
	t.Parallel()
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db, err := testpostgres.New(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Pool.Exec(tenantCtx, "alter table pages_content add constraint reject_page_title check (title <> 'Rejected')"); err != nil {
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
	err = postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
		contentID := uuid.New()
		if err := repository.CreatePage(ctx, item, contentID); err != nil {
			return err
		}
		return repository.CreatePageContent(ctx, item.ID, contentID, item.Title, item.HTML, *item.CreatedAt)
	})
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != pgerrcode.CheckViolation {
		t.Fatalf("error=%v, want rejected content version", err)
	}

	var pagesCount, versions int
	if err := db.Pool.QueryRow(tenantCtx, "select (select count(*) from pages), (select count(*) from pages_content)").Scan(&pagesCount, &versions); err != nil {
		t.Fatal(err)
	}
	if pagesCount != 0 || versions != 0 {
		t.Errorf("failed create left %d pages and %d versions", pagesCount, versions)
	}
}

func TestPagesRepositoryDeletePage(t *testing.T) {
	t.Parallel()
	tenantCtx := tenant.WithKey(t.Context(), tenant.Production())
	db, err := testpostgres.New(tenantCtx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	_, err = db.Pool.Exec(tenantCtx, `
		insert into pages (
			tenant, id, namespace, slug, current_content_id, published_at, created_at, updated_at
		)
		values
			('tadoku/prod', '11111111-1111-4111-8111-111111111111',
			 'main', 'first-page', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
			 '2026-09-11 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00'),
			('tadoku/prod', '22222222-2222-4222-8222-222222222222',
			 'main', 'second-page', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1',
			 null, '2026-09-10 13:00:00', '2026-09-10 13:00:00');
		insert into pages_content (tenant, id, page_id, title, html, created_at)
		values
			('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111',
			 'Original title', 'Original html', '2026-09-10 12:00:00'),
			('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111',
			 'Revised title', 'Revised html', '2026-09-11 12:00:00'),
			('tadoku/prod', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '22222222-2222-4222-8222-222222222222',
			 'Other page', 'Other html', '2026-09-10 13:00:00');`)
	if err != nil {
		t.Fatal(err)
	}

	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	deletedAt := time.Date(2026, 9, 12, 21, 0, 0, 0, time.FixedZone("UTC+9", 9*60*60))
	repository := pages.NewPagesRepository(db.Pool)
	const postSnapshotSQL = `select (to_jsonb(pages) - 'deleted_at')::text, deleted_at
		from pages where id = $1`
	const versionsSnapshotSQL = `select jsonb_agg(to_jsonb(pages_content) order by id)::text from pages_content`

	var before, versionsBefore string
	var initialDeletedAt *time.Time
	if err := db.Pool.QueryRow(tenantCtx, postSnapshotSQL, id).Scan(&before, &initialDeletedAt); err != nil {
		t.Fatal(err)
	}
	if initialDeletedAt != nil {
		t.Fatal("seed page is already deleted")
	}
	if err := db.Pool.QueryRow(tenantCtx, versionsSnapshotSQL).Scan(&versionsBefore); err != nil {
		t.Fatal(err)
	}

	if err := repository.DeletePage(tenantCtx, "other", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	var wrongNamespaceDeletedAt *time.Time
	if err := db.Pool.QueryRow(tenantCtx, "select deleted_at from pages where id = $1", id).Scan(&wrongNamespaceDeletedAt); err != nil {
		t.Fatal(err)
	}
	if wrongNamespaceDeletedAt != nil {
		t.Fatal("wrong namespace deleted the page")
	}

	if err := repository.DeletePage(tenantCtx, "main", uuid.MustParse("99999999-9999-4999-8999-999999999999"), deletedAt); err != nil {
		t.Fatalf("missing page must be an idempotent success: %v", err)
	}

	if err := repository.DeletePage(tenantCtx, "main", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.DeletePage(tenantCtx, "main", id, deletedAt.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	var after, versionsAfter string
	var gotDeletedAt time.Time
	if err := db.Pool.QueryRow(tenantCtx, postSnapshotSQL, id).Scan(&after, &gotDeletedAt); err != nil {
		t.Fatalf("soft-deleted row must still exist: %v", err)
	}
	if after != before {
		t.Errorf("deletion changed fields other than deleted_at:\nbefore %s\nafter %s", before, after)
	}
	if !gotDeletedAt.Equal(deletedAt) {
		t.Errorf("deleted_at=%v, want original timestamp %v", gotDeletedAt, deletedAt)
	}

	if err := db.Pool.QueryRow(tenantCtx, versionsSnapshotSQL).Scan(&versionsAfter); err != nil {
		t.Fatal(err)
	}
	if versionsAfter != versionsBefore {
		t.Errorf("deletion changed html versions:\nbefore %s\nafter %s", versionsBefore, versionsAfter)
	}

	var otherDeletedAt *time.Time
	if err := db.Pool.QueryRow(tenantCtx, "select deleted_at from pages where id = $1", "22222222-2222-4222-8222-222222222222").Scan(&otherDeletedAt); err != nil {
		t.Fatalf("other page must remain present: %v", err)
	}
	if otherDeletedAt != nil {
		t.Errorf("other page was deleted at %v", otherDeletedAt)
	}
}
