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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestPagesRepositoryReadsCurrentLiveContent(t *testing.T) {
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

	_, err = db.Pool.Exec(t.Context(), `
		insert into pages (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
		values
			('10000000-0000-4000-8000-000000000001', 'main', 'first-page', '20000000-0000-4000-8000-000000000002', '2026-09-12 12:00:01', '2026-09-12 11:00:00', '2026-09-12 11:30:00', null),
			('10000000-0000-4000-8000-000000000002', 'main', 'draft-page', '20000000-0000-4000-8000-000000000003', null, '2026-09-12 11:00:00', '2026-09-12 11:00:00', null),
			('10000000-0000-4000-8000-000000000003', 'other', 'other-page', '20000000-0000-4000-8000-000000000004', '2026-09-12 10:00:00', '2026-09-12 11:00:00', '2026-09-12 11:00:00', null),
			('10000000-0000-4000-8000-000000000004', 'main', 'deleted-page', '20000000-0000-4000-8000-000000000005', '2026-09-12 10:00:00', '2026-09-12 12:00:00', '2026-09-12 12:00:00', '2026-09-12 12:30:00');
		insert into pages_content (id, page_id, title, html, created_at)
		values
			('20000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000001', 'Old title', '<p>Old</p>', '2026-09-12 10:00:00'),
			('20000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000001', 'Current title', '<p>Current</p>', '2026-09-12 11:30:00'),
			('20000000-0000-4000-8000-000000000003', '10000000-0000-4000-8000-000000000002', 'Draft', '<p>Draft</p>', '2026-09-12 11:00:00'),
			('20000000-0000-4000-8000-000000000004', '10000000-0000-4000-8000-000000000003', 'Other', '<p>Other</p>', '2026-09-12 11:00:00'),
			('20000000-0000-4000-8000-000000000005', '10000000-0000-4000-8000-000000000004', 'Deleted', '<p>Deleted</p>', '2026-09-12 12:00:00')`)
	if err != nil {
		t.Fatal(err)
	}

	repository := pages.NewPagesRepository(db.Pool)
	page, err := repository.FindPageBySlug(t.Context(), "main", "first-page")
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
		if _, err := repository.FindPageBySlug(t.Context(), lookup.namespace, lookup.slug); !errors.Is(err, pages.ErrPageNotFound) {
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
			items, total, err := repository.ListPages(t.Context(), "main", test.includeDrafts, test.cutoff, test.limit, test.offset)
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
	if err := db.Pool.QueryRow(t.Context(), "select (select count(*) from pages), (select count(*) from pages_content)").Scan(&pagesCount, &versions); err != nil {
		t.Fatal(err)
	}
	if pagesCount != 1 || versions != 1 {
		t.Errorf("created %d pages and %d versions, want one each", pagesCount, versions)
	}

	duplicate := *item
	duplicate.ID = uuid.New()
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.CreatePage(ctx, &duplicate, uuid.New())
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
	if err := db.Pool.QueryRow(t.Context(), "select (select count(*) from pages), (select count(*) from pages_content)").Scan(&pagesCount, &versions); err != nil {
		t.Fatal(err)
	}
	if pagesCount != 0 || versions != 0 {
		t.Errorf("failed create left %d pages and %d versions", pagesCount, versions)
	}
}

func TestPagesRepositoryDeletePage(t *testing.T) {
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

	_, err = db.Pool.Exec(t.Context(), `
		insert into pages (id, namespace, slug, current_content_id, published_at, created_at, updated_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'first-page', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '2026-09-11 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00'),
			('22222222-2222-4222-8222-222222222222', 'main', 'second-page', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00');
		insert into pages_content (id, page_id, title, html, created_at)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'Original title', 'Original html', '2026-09-10 12:00:00'),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', 'Revised title', 'Revised html', '2026-09-11 12:00:00'),
			('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '22222222-2222-4222-8222-222222222222', 'Other page', 'Other html', '2026-09-10 13:00:00');`)
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
	if err := db.Pool.QueryRow(t.Context(), postSnapshotSQL, id).Scan(&before, &initialDeletedAt); err != nil {
		t.Fatal(err)
	}
	if initialDeletedAt != nil {
		t.Fatal("seed page is already deleted")
	}
	if err := db.Pool.QueryRow(t.Context(), versionsSnapshotSQL).Scan(&versionsBefore); err != nil {
		t.Fatal(err)
	}

	if err := repository.DeletePage(t.Context(), "other", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	var wrongNamespaceDeletedAt *time.Time
	if err := db.Pool.QueryRow(t.Context(), "select deleted_at from pages where id = $1", id).Scan(&wrongNamespaceDeletedAt); err != nil {
		t.Fatal(err)
	}
	if wrongNamespaceDeletedAt != nil {
		t.Fatal("wrong namespace deleted the page")
	}
	if err := repository.DeletePage(t.Context(), "main", uuid.MustParse("99999999-9999-4999-8999-999999999999"), deletedAt); err != nil {
		t.Fatalf("missing page must be an idempotent success: %v", err)
	}
	if err := repository.DeletePage(t.Context(), "main", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.DeletePage(t.Context(), "main", id, deletedAt.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	var after, versionsAfter string
	var gotDeletedAt time.Time
	if err := db.Pool.QueryRow(t.Context(), postSnapshotSQL, id).Scan(&after, &gotDeletedAt); err != nil {
		t.Fatalf("soft-deleted row must still exist: %v", err)
	}
	if after != before {
		t.Errorf("deletion changed fields other than deleted_at:\nbefore %s\nafter %s", before, after)
	}
	if !gotDeletedAt.Equal(deletedAt) {
		t.Errorf("deleted_at=%v, want original timestamp %v", gotDeletedAt, deletedAt)
	}
	if err := db.Pool.QueryRow(t.Context(), versionsSnapshotSQL).Scan(&versionsAfter); err != nil {
		t.Fatal(err)
	}
	if versionsAfter != versionsBefore {
		t.Errorf("deletion changed html versions:\nbefore %s\nafter %s", versionsBefore, versionsAfter)
	}
	var otherDeletedAt *time.Time
	if err := db.Pool.QueryRow(t.Context(), "select deleted_at from pages where id = $1", "22222222-2222-4222-8222-222222222222").Scan(&otherDeletedAt); err != nil {
		t.Fatalf("other page must remain present: %v", err)
	}
	if otherDeletedAt != nil {
		t.Errorf("other page was deleted at %v", otherDeletedAt)
	}
}

func TestPagesRepositoryUpdatePage(t *testing.T) {
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

	_, err = db.Pool.Exec(t.Context(), `
		insert into pages (id, namespace, slug, current_content_id, published_at, created_at, updated_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'original-page', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', null, '2026-09-10 12:00:00', '2026-09-10 12:00:00'),
			('22222222-2222-4222-8222-222222222222', 'main', 'other-page', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00');
		insert into pages_content (id, page_id, title, html, created_at)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '11111111-1111-4111-8111-111111111111', 'Original title', 'Original html', '2026-09-10 12:00:00'),
			('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', '22222222-2222-4222-8222-222222222222', 'Other title', 'Other html', '2026-09-10 13:00:00');
		alter table pages_content add constraint reject_update_title check (title <> 'rejected revision');`)
	if err != nil {
		t.Fatal(err)
	}

	repository := pages.NewPagesRepository(db.Pool)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	original, err := repository.FindPageByID(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}

	metadata := *original
	metadata.Slug = "renamed-page"
	publishedAt := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	metadata.PublishedAt = &publishedAt
	metadata.UpdatedAt = &publishedAt
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.UpdatePage(ctx, &metadata, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repository.FindPageByID(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, &metadata) {
		t.Errorf("metadata update=%+v, want %+v", got, metadata)
	}
	var contentID uuid.UUID
	var count int
	if err := db.Pool.QueryRow(t.Context(), `select current_content_id, (select count(*) from pages_content where page_id = $1) from pages where id = $1`, id).Scan(&contentID, &count); err != nil {
		t.Fatal(err)
	}
	if contentID != uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa") || count != 1 {
		t.Errorf("metadata update changed revision: current=%v, count=%d", contentID, count)
	}

	updated := metadata
	updated.Title = "Revised title"
	updated.HTML = "Revised html"
	updated.PublishedAt = nil
	updatedAt := publishedAt.Add(time.Hour)
	updated.UpdatedAt = &updatedAt
	stop := errors.New("roll back revised page")
	for _, rollback := range []bool{true, false} {
		err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
			contentID := uuid.New()
			if err := repository.UpdatePage(ctx, &updated, &contentID); err != nil {
				return err
			}
			if err := repository.CreatePageContent(ctx, updated.ID, contentID, updated.Title, updated.HTML, *updated.UpdatedAt); err != nil {
				return err
			}
			got, err := repository.FindPageByID(ctx, "main", id)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(got, &updated) {
				t.Errorf("read within transaction=%+v, want %+v", got, updated)
			}
			outside, err := repository.FindPageByID(t.Context(), "main", id)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(outside, &metadata) {
				t.Errorf("uncommitted revision escaped transaction: %+v", outside)
			}
			if rollback {
				return stop
			}
			return nil
		})
		want := &updated
		wantCount := 2
		if rollback {
			want = &metadata
			wantCount = 1
			if !errors.Is(err, stop) {
				t.Fatalf("rollback error=%v, want %v", err, stop)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		got, err := repository.FindPageByID(t.Context(), "main", id)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("rollback=%t: persisted page=%+v, want %+v", rollback, got, want)
		}
		if err := db.Pool.QueryRow(t.Context(), "select count(*) from pages_content where page_id = $1", id).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != wantCount {
			t.Errorf("rollback=%t: revision count=%d, want %d", rollback, count, wantCount)
		}
	}

	var title, body string
	var createdAt time.Time
	if err := db.Pool.QueryRow(t.Context(), `select title, html, created_at from pages_content where id = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa'`).Scan(&title, &body, &createdAt); err != nil {
		t.Fatal(err)
	}
	if title != original.Title || body != original.HTML || !createdAt.Equal(*original.CreatedAt) {
		t.Errorf("original revision changed: title=%q, html=%q, created_at=%v", title, body, createdAt)
	}
	if err := db.Pool.QueryRow(t.Context(), `select pages_content.created_at from pages join pages_content on pages_content.id = pages.current_content_id where pages.id = $1`, id).Scan(&createdAt); err != nil {
		t.Fatal(err)
	}
	if !createdAt.Equal(*updated.UpdatedAt) {
		t.Errorf("revision timestamp=%v, want explicit update time %v", createdAt, updated.UpdatedAt)
	}

	const snapshotSQL = `select jsonb_build_object(
		'pages', (select jsonb_agg(to_jsonb(p) order by id) from pages p),
		'html', (select jsonb_agg(to_jsonb(c) order by id) from pages_content c)
	)::text`
	for _, test := range []struct {
		name   string
		change func(*pages.Page)
		want   error
	}{
		{name: "wrong namespace", change: func(page *pages.Page) { page.Namespace = "other" }, want: pages.ErrPageNotFound},
		{name: "missing page", change: func(page *pages.Page) { page.ID = uuid.MustParse("99999999-9999-4999-8999-999999999999") }, want: pages.ErrPageNotFound},
		{name: "duplicate slug", change: func(page *pages.Page) { page.Slug = "other-page" }, want: pages.ErrPageAlreadyExists},
		{name: "revision insertion failure", change: func(page *pages.Page) {
			page.Slug = "rolled-back-page"
			page.Title = "rejected revision"
			updatedAt := page.UpdatedAt.Add(time.Hour)
			page.UpdatedAt = &updatedAt
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var before, after string
			if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&before); err != nil {
				t.Fatal(err)
			}
			attempt := updated
			test.change(&attempt)
			err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
				contentID := uuid.New()
				if err := repository.UpdatePage(ctx, &attempt, &contentID); err != nil {
					return err
				}
				return repository.CreatePageContent(ctx, attempt.ID, contentID, attempt.Title, attempt.HTML, *attempt.UpdatedAt)
			})
			if test.want != nil {
				if !errors.Is(err, test.want) {
					t.Fatalf("error=%v, want %v", err, test.want)
				}
			} else {
				var pgError *pgconn.PgError
				if !errors.As(err, &pgError) || pgError.Code != pgerrcode.CheckViolation || pgError.ConstraintName != "reject_update_title" {
					t.Fatalf("error=%v, want revision constraint failure", err)
				}
			}
			if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if before != after {
				t.Errorf("failed update changed storage:\nbefore %s\nafter %s", before, after)
			}
		})
	}

	if _, err := db.Pool.Exec(t.Context(), "update pages set deleted_at = $1 where id = $2", *updated.UpdatedAt, id); err != nil {
		t.Fatal(err)
	}
	var before, after string
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&before); err != nil {
		t.Fatal(err)
	}
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		contentID := uuid.New()
		if err := repository.UpdatePage(ctx, &updated, &contentID); err != nil {
			return err
		}
		return repository.CreatePageContent(ctx, updated.ID, contentID, updated.Title, updated.HTML, *updated.UpdatedAt)
	})
	if !errors.Is(err, pages.ErrPageNotFound) {
		t.Errorf("deleted page error=%v, want not found", err)
	}
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("deleted page update changed storage:\nbefore %s\nafter %s", before, after)
	}
}

func TestPagesRepositoryGetPageVersion(t *testing.T) {
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

	_, err = db.Pool.Exec(t.Context(), `
		insert into pages (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'published-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '2026-09-11 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00', null),
			('22222222-2222-4222-8222-222222222222', 'main', 'draft-post', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00', null),
			('33333333-3333-4333-8333-333333333333', 'main', 'deleted-post', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00', '2026-09-11 12:00:00');
		insert into pages_content (id, page_id, title, html, created_at)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '11111111-1111-4111-8111-111111111111', 'Latest title', 'Latest content', '2026-09-11 12:00:00'),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', 'Tied title', 'Tied content', '2026-09-10 12:00:00'),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'Original title', 'Original content', '2026-09-10 12:00:00'),
			('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '22222222-2222-4222-8222-222222222222', 'Draft title', 'Draft content', '2026-09-10 13:00:00'),
			('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '33333333-3333-4333-8333-333333333333', 'Deleted title', 'Deleted content', '2026-09-10 13:00:00'),
			('dddddddd-dddd-4ddd-8ddd-ddddddddddd1', '44444444-4444-4444-8444-444444444444', 'Orphan title', 'Orphan content', '2026-09-10 13:00:00');`)
	if err != nil {
		t.Fatal(err)
	}

	repository := pages.NewPagesRepository(db.Pool)
	pageID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	for _, test := range []struct {
		id        string
		version   int
		title     string
		body      string
		createdAt time.Time
	}{
		{"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1", 1, "Original title", "Original content", time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)},
		{"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2", 2, "Tied title", "Tied content", time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)},
		{"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3", 3, "Latest title", "Latest content", time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)},
	} {
		t.Run(test.title, func(t *testing.T) {
			id := uuid.MustParse(test.id)
			got, err := repository.GetPageVersion(t.Context(), "main", pageID, id)
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != id || got.Version != test.version || got.Title != test.title || got.HTML != test.body || !got.CreatedAt.Equal(test.createdAt) {
				t.Errorf("version=%+v, want ID=%s ordinal=%d title=%q content=%q created_at=%v", got, id, test.version, test.title, test.body, test.createdAt)
			}
		})
	}

	for _, test := range []struct {
		name      string
		namespace string
		pageID    string
		contentID string
	}{
		{"wrong namespace", "other", pageID.String(), "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"},
		{"wrong page", "main", "22222222-2222-4222-8222-222222222222", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"},
		{"missing revision", "main", pageID.String(), "99999999-9999-4999-8999-999999999999"},
		{"deleted page", "main", "33333333-3333-4333-8333-333333333333", "cccccccc-cccc-4ccc-8ccc-ccccccccccc1"},
		{"missing parent", "main", "44444444-4444-4444-8444-444444444444", "dddddddd-dddd-4ddd-8ddd-ddddddddddd1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := repository.GetPageVersion(t.Context(), test.namespace, uuid.MustParse(test.pageID), uuid.MustParse(test.contentID))
			if !errors.Is(err, pages.ErrPageNotFound) {
				t.Errorf("error=%v, want page not found", err)
			}
		})
	}
}

func TestPagesRepositoryListPageVersions(t *testing.T) {
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

	_, err = db.Pool.Exec(t.Context(), `
		insert into pages (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'scheduled', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '2026-09-14 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00', null),
			('22222222-2222-4222-8222-222222222222', 'main', 'deleted', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', null, '2026-09-10 12:00:00', '2026-09-11 12:00:00', '2026-09-12 12:00:00');
		insert into pages_content (id, page_id, title, html, created_at)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '11111111-1111-4111-8111-111111111111', 'Third', 'Third body', '2026-09-11 12:00:00'),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', 'Second', 'Second body', '2026-09-10 12:00:00'),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'First', 'First body', '2026-09-10 12:00:00'),
			('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '22222222-2222-4222-8222-222222222222', 'Deleted', 'Deleted body', '2026-09-10 12:00:00'),
			('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '33333333-3333-4333-8333-333333333333', 'Orphan', 'Orphan body', '2026-09-10 12:00:00');`)
	if err != nil {
		t.Fatal(err)
	}

	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	firstCreatedAt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	want := []pages.PageVersion{
		{
			ID:        uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"),
			Version:   1,
			Title:     "First",
			CreatedAt: firstCreatedAt,
		},
		{
			ID:        uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2"),
			Version:   2,
			Title:     "Second",
			CreatedAt: firstCreatedAt,
		},
		{
			ID:        uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3"),
			Version:   3,
			Title:     "Third",
			CreatedAt: firstCreatedAt.Add(24 * time.Hour),
		},
	}
	repository := pages.NewPagesRepository(db.Pool)
	versions, err := repository.ListPageVersions(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(versions, want) {
		t.Errorf("versions=%+v, want %+v", versions, want)
	}

	for _, test := range []struct {
		name      string
		namespace string
		id        uuid.UUID
	}{
		{name: "wrong namespace", namespace: "other", id: id},
		{name: "deleted", namespace: "main", id: uuid.MustParse("22222222-2222-4222-8222-222222222222")},
		{name: "orphan", namespace: "main", id: uuid.MustParse("33333333-3333-4333-8333-333333333333")},
		{name: "missing", namespace: "main", id: uuid.MustParse("44444444-4444-4444-8444-444444444444")},
	} {
		t.Run(test.name, func(t *testing.T) {
			versions, err := repository.ListPageVersions(t.Context(), test.namespace, test.id)
			if err != nil {
				t.Fatal(err)
			}
			if len(versions) != 0 {
				t.Errorf("versions=%+v, want empty history", versions)
			}
		})
	}

	wantRollback := errors.New("rollback history edit")
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		executor, err := postgres.Executor(ctx, db.Pool)
		if err != nil {
			return err
		}
		if _, err := executor.Exec(ctx, "update pages_content set title = 'Uncommitted' where id = $1", want[0].ID); err != nil {
			return err
		}

		versions, err := repository.ListPageVersions(ctx, "main", id)
		if err != nil {
			return err
		}
		if len(versions) != 3 || versions[0].Title != "Uncommitted" {
			t.Errorf("transaction history=%+v", versions)
		}
		return wantRollback
	})
	if !errors.Is(err, wantRollback) {
		t.Fatalf("transaction error=%v, want rollback", err)
	}
	versions, err = repository.ListPageVersions(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(versions, want) {
		t.Errorf("rolled-back history=%+v, want %+v", versions, want)
	}
}
