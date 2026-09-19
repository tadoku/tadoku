package pages_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
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

	for _, test := range []struct {
		name          string
		includeDrafts bool
		limit         int32
		offset        int64
		wantIDs       []uuid.UUID
		wantTotal     int
	}{
		{name: "all live rows with stable ties", includeDrafts: true, limit: 10, wantIDs: []uuid.UUID{
			uuid.MustParse("10000000-0000-4000-8000-000000000002"),
			uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		}, wantTotal: 2},
		{name: "scheduled is not a draft", limit: 10, wantIDs: []uuid.UUID{
			uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		}, wantTotal: 1},
		{name: "empty page preserves total", includeDrafts: true, limit: 10, offset: 100, wantIDs: []uuid.UUID{}, wantTotal: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			items, total, err := repository.ListPages(t.Context(), "main", test.includeDrafts, test.limit, test.offset)
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
