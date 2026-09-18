package pages_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

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
