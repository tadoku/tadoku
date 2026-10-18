package posts_test

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestPostsRepositoryFindUsesCurrentContentAndNamespace(t *testing.T) {
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
		insert into posts (
			tenant, id, namespace, slug, current_content_id, published_at, created_at, updated_at
		)
		values
			('tadoku/prod', '11111111-1111-4111-8111-111111111111',
			 'main', 'welcome', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
			 '2026-09-11 10:00:00', '2026-09-10 08:00:00', '2026-09-11 09:30:00'),
			('tadoku/prod', '22222222-2222-4222-8222-222222222222',
			 'other', 'welcome', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',
			 null, '2026-09-10 08:00:00', '2026-09-11 09:30:00');
		insert into posts_content (tenant, id, post_id, title, content, created_at)
		values
			('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '11111111-1111-4111-8111-111111111111',
			 'Current title', 'Current content', '2026-09-11 09:30:00'),
			('tadoku/prod', 'cccccccc-cccc-4ccc-8ccc-cccccccccccc', '11111111-1111-4111-8111-111111111111',
			 'Old title', 'Old content', '2026-09-10 08:00:00'),
			('tadoku/prod', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', '22222222-2222-4222-8222-222222222222',
			 'Other title', 'Other content', '2026-09-11 09:30:00')`)
	if err != nil {
		t.Fatal(err)
	}

	repository := posts.NewPostsRepository(db.Pool)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")

	publishedAt := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 9, 11, 9, 30, 0, 0, time.UTC)
	want := &posts.Post{
		ID:          id,
		Namespace:   "main",
		Slug:        "welcome",
		Title:       "Current title",
		Content:     "Current content",
		PublishedAt: &publishedAt,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}

	for name, find := range map[string]func() (*posts.Post, error){
		"slug": func() (*posts.Post, error) { return repository.FindPostBySlug(tenantCtx, "main", "welcome") },
		"ID":   func() (*posts.Post, error) { return repository.FindPostByID(tenantCtx, "main", id) },
	} {
		got, err := find()
		if err != nil {
			t.Fatalf("%s lookup: %v", name, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s lookup=%+v, want %+v", name, got, want)
		}
	}

	other, err := repository.FindPostBySlug(tenantCtx, "other", "welcome")
	if err != nil {
		t.Fatal(err)
	}
	if other.Title != "Other title" || other.PublishedAt != nil {
		t.Errorf("other namespace draft=%+v", other)
	}
	if _, err := repository.FindPostByID(tenantCtx, "other", id); !errors.Is(err, posts.ErrPostNotFound) {
		t.Errorf("wrong namespace error=%v, want post not found", err)
	}

	wantRollback := errors.New("roll back post read")
	err = postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
		executor, err := postgres.Executor(ctx, db.Pool)
		if err != nil {
			return err
		}
		if _, err := executor.Exec(ctx, "update posts set current_content_id = 'cccccccc-cccc-4ccc-8ccc-cccccccccccc' where id = $1", id); err != nil {
			return err
		}

		post, err := repository.FindPostBySlug(ctx, "main", "welcome")
		if err != nil {
			return err
		}
		if post.Title != "Old title" {
			t.Errorf("slug lookup did not use transaction: %+v", post)
		}

		post, err = repository.FindPostByID(ctx, "main", id)
		if err != nil {
			return err
		}
		if post.Content != "Old content" {
			t.Errorf("ID lookup did not use transaction: %+v", post)
		}
		return wantRollback
	})
	if !errors.Is(err, wantRollback) {
		t.Fatal(err)
	}

	got, err := repository.FindPostByID(tenantCtx, "main", id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rollback changed persisted post: %+v", got)
	}

	if _, err := db.Pool.Exec(tenantCtx, "update posts set deleted_at = '2026-09-12 12:00:00' where id = $1", id); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindPostBySlug(tenantCtx, "main", "welcome"); !errors.Is(err, posts.ErrPostNotFound) {
		t.Errorf("deleted slug error=%v, want post not found", err)
	}
	if _, err := repository.FindPostByID(tenantCtx, "main", id); !errors.Is(err, posts.ErrPostNotFound) {
		t.Errorf("deleted ID error=%v, want post not found", err)
	}
}

func TestPostsRepositoryListPosts(t *testing.T) {
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
		insert into posts (
			tenant, id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at
		)
		values
			('tadoku/prod', '10000000-0000-4000-8000-000000000001', 'main',
			 'older', '20000000-0000-4000-8000-000000000001', '2026-09-11 12:00:00',
			 '2026-09-10 10:00:00', '2026-09-11 12:00:00', null),
			('tadoku/prod', '10000000-0000-4000-8000-000000000002', 'main',
			 'boundary', '20000000-0000-4000-8000-000000000002', '2026-09-12 12:00:00',
			 '2026-09-11 10:00:00', '2026-09-12 11:00:00', null),
			('tadoku/prod', '10000000-0000-4000-8000-000000000003', 'main',
			 'scheduled', '20000000-0000-4000-8000-000000000003', '2026-09-12 12:00:01',
			 '2026-09-11 11:00:00', '2026-09-12 11:00:00', null),
			('tadoku/prod', '10000000-0000-4000-8000-000000000004', 'main',
			 'draft', '20000000-0000-4000-8000-000000000004', null,
			 '2026-09-11 12:00:00', '2026-09-12 11:00:00', null),
			('tadoku/prod', '10000000-0000-4000-8000-000000000005', 'main',
			 'deleted', '20000000-0000-4000-8000-000000000005', '2026-09-11 12:00:00',
			 '2026-09-11 13:00:00', '2026-09-12 11:00:00', '2026-09-12 11:00:00'),
			('tadoku/prod', '10000000-0000-4000-8000-000000000006', 'other',
			 'other', '20000000-0000-4000-8000-000000000006', '2026-09-11 12:00:00',
			 '2026-09-11 14:00:00', '2026-09-12 11:00:00', null),
			('tadoku/prod', '10000000-0000-4000-8000-000000000007', 'main',
			 'tied', '20000000-0000-4000-8000-000000000007', '2026-09-12 12:00:00',
			 '2026-09-11 10:00:00', '2026-09-12 11:00:00', null);
		insert into posts_content (tenant, id, post_id, title, content, created_at)
		values
			('tadoku/prod', '20000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000001',
			 'Older', 'Older content', '2026-09-10 10:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000002',
			 'Boundary', 'Boundary content', '2026-09-11 10:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000003', '10000000-0000-4000-8000-000000000003',
			 'Scheduled', 'Scheduled content', '2026-09-11 11:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000004', '10000000-0000-4000-8000-000000000004',
			 'Draft', 'Draft content', '2026-09-11 12:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000005', '10000000-0000-4000-8000-000000000005',
			 'Deleted', 'Deleted content', '2026-09-11 13:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000006', '10000000-0000-4000-8000-000000000006',
			 'Other', 'Other content', '2026-09-11 14:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000007', '10000000-0000-4000-8000-000000000007',
			 'Current title', 'Current content', '2026-09-12 11:00:00'),
			('tadoku/prod', '20000000-0000-4000-8000-000000000008', '10000000-0000-4000-8000-000000000007',
			 'Old title', 'Old content', '2026-09-11 10:00:00');`)
	if err != nil {
		t.Fatal(err)
	}

	cutoff := time.Date(2026, 9, 12, 21, 0, 0, 0, time.FixedZone("UTC+9", 9*60*60))
	repository := posts.NewPostsRepository(db.Pool)

	tests := []struct {
		name          string
		namespace     string
		includeDrafts bool
		cutoff        time.Time
		limit         int32
		offset        int64
		wantIDs       []byte
		wantTotal     int
	}{
		{name: "published by cutoff", namespace: "main", cutoff: cutoff, limit: 10, wantIDs: []byte{7, 2, 1}, wantTotal: 3},
		{name: "before cutoff", namespace: "main", cutoff: cutoff.Add(-time.Second), limit: 10, wantIDs: []byte{1}, wantTotal: 1},
		{name: "include drafts and scheduled", namespace: "main", includeDrafts: true, cutoff: cutoff, limit: 10, wantIDs: []byte{4, 3, 7, 2, 1}, wantTotal: 5},
		{name: "other namespace", namespace: "other", cutoff: cutoff, limit: 10, wantIDs: []byte{6}, wantTotal: 1},
		{name: "missing namespace", namespace: "missing", cutoff: cutoff, limit: 10, wantIDs: []byte{}},
		{name: "first tied page", namespace: "main", cutoff: cutoff, limit: 1, wantIDs: []byte{7}, wantTotal: 3},
		{name: "second tied page", namespace: "main", cutoff: cutoff, limit: 1, offset: 1, wantIDs: []byte{2}, wantTotal: 3},
		{name: "empty page preserves total", namespace: "main", cutoff: cutoff, limit: 10, offset: math.MaxInt64, wantIDs: []byte{}, wantTotal: 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items, total, err := repository.ListPosts(tenantCtx, test.namespace, test.includeDrafts, test.cutoff, test.limit, test.offset)
			if err != nil {
				t.Fatal(err)
			}
			if total != test.wantTotal {
				t.Errorf("total=%d, want %d", total, test.wantTotal)
			}

			wantIDs := make([]uuid.UUID, 0, len(test.wantIDs))
			for _, suffix := range test.wantIDs {
				id := uuid.MustParse("10000000-0000-4000-8000-000000000000")
				id[15] = suffix
				wantIDs = append(wantIDs, id)
			}

			gotIDs := make([]uuid.UUID, 0, len(items))
			for _, item := range items {
				gotIDs = append(gotIDs, item.ID)
			}
			if !reflect.DeepEqual(gotIDs, wantIDs) {
				t.Fatalf("IDs=%v, want %v", gotIDs, wantIDs)
			}

			for _, item := range items {
				if item.ID[15] == 4 && item.PublishedAt != nil {
					t.Errorf("draft publication=%v, want nil", item.PublishedAt)
				}
				if item.ID[15] != 7 {
					continue
				}

				publishedAt := cutoff.UTC()
				createdAt := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
				updatedAt := time.Date(2026, 9, 12, 11, 0, 0, 0, time.UTC)
				want := posts.Post{
					ID:          item.ID,
					Namespace:   "main",
					Slug:        "tied",
					Title:       "Current title",
					Content:     "Current content",
					PublishedAt: &publishedAt,
					CreatedAt:   &createdAt,
					UpdatedAt:   &updatedAt,
				}

				if !reflect.DeepEqual(item, want) {
					t.Errorf("post=%+v, want %+v", item, want)
				}
			}
		})
	}
}
