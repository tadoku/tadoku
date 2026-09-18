package pages_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

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
