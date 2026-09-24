package posts_test

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestPostsRepositoryCreatePost(t *testing.T) {
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

	instant := time.Date(2026, 9, 12, 21, 0, 0, 0, time.FixedZone("UTC+9", 9*60*60))
	repository := posts.NewPostsRepository(db.Pool)
	for _, test := range []struct {
		name        string
		publishedAt *time.Time
	}{
		{name: "draft"},
		{name: "published", publishedAt: &instant},
	} {
		t.Run(test.name, func(t *testing.T) {
			createdAt := instant.Add(-2 * time.Hour)
			updatedAt := instant.Add(-time.Hour)
			item := &posts.Post{
				ID:          uuid.New(),
				Namespace:   "main",
				Slug:        test.name,
				Title:       "Title",
				Content:     "Content",
				PublishedAt: test.publishedAt,
				CreatedAt:   &createdAt,
				UpdatedAt:   &updatedAt,
			}
			want := *item
			createdAtUTC := want.CreatedAt.UTC()
			updatedAtUTC := want.UpdatedAt.UTC()
			want.CreatedAt = &createdAtUTC
			want.UpdatedAt = &updatedAtUTC
			if want.PublishedAt != nil {
				utc := want.PublishedAt.UTC()
				want.PublishedAt = &utc
			}

			err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
				contentID := uuid.New()
				if err := repository.CreatePost(ctx, item, contentID); err != nil {
					return err
				}
				if err := repository.CreatePostContent(ctx, item.ID, contentID, item.Title, item.Content, *item.CreatedAt); err != nil {
					return err
				}
				got, err := repository.FindPostByID(ctx, item.Namespace, item.ID)
				if err != nil {
					return err
				}
				if !reflect.DeepEqual(got, &want) {
					t.Errorf("transaction readback=%+v, want %+v", got, want)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			got, err := repository.FindPostByID(t.Context(), item.Namespace, item.ID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, &want) {
				t.Errorf("committed post=%+v, want %+v", got, want)
			}

			var contentID uuid.UUID
			var contentCreatedAt time.Time
			if err := db.Pool.QueryRow(t.Context(), `
				select posts_content.id, posts_content.created_at
				from posts
				join posts_content on posts_content.id = posts.current_content_id
				where posts.id = $1 and posts_content.post_id = posts.id`, item.ID,
			).Scan(&contentID, &contentCreatedAt); err != nil {
				t.Fatal(err)
			}
			if contentID == uuid.Nil {
				t.Error("first revision has a zero ID")
			}
			if !contentCreatedAt.Equal(*item.CreatedAt) {
				t.Errorf("revision timestamp=%v, want %v", contentCreatedAt, item.CreatedAt)
			}
			var revisions int
			if err := db.Pool.QueryRow(t.Context(), "select count(*) from posts_content where post_id = $1", item.ID).Scan(&revisions); err != nil {
				t.Fatal(err)
			}
			if revisions != 1 {
				t.Errorf("created %d revisions, want one", revisions)
			}
		})
	}
}

func TestPostsRepositoryCreatePostConflictsPreserveData(t *testing.T) {
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

	repository := posts.NewPostsRepository(db.Pool)
	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	original := &posts.Post{
		ID:        uuid.New(),
		Namespace: "main",
		Slug:      "first-post",
		Title:     "Original title",
		Content:   "Original content",
		CreatedAt: &instant,
		UpdatedAt: &instant,
	}
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		contentID := uuid.New()
		if err := repository.CreatePost(ctx, original, contentID); err != nil {
			return err
		}

		return repository.CreatePostContent(ctx, original.ID, contentID, original.Title, original.Content, *original.CreatedAt)
	}); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name      string
		id        uuid.UUID
		namespace string
		slug      string
	}{
		{name: "global ID", id: original.ID, namespace: "other", slug: "other-slug"},
		{name: "namespace slug", id: uuid.New(), namespace: "main", slug: original.Slug},
	} {
		t.Run(test.name, func(t *testing.T) {
			duplicate := *original
			duplicate.ID = test.id
			duplicate.Namespace = test.namespace
			duplicate.Slug = test.slug
			duplicate.Title = "Must not replace the original"
			err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
				contentID := uuid.New()
				if err := repository.CreatePost(ctx, &duplicate, contentID); err != nil {
					return err
				}

				return repository.CreatePostContent(ctx, duplicate.ID, contentID, duplicate.Title, duplicate.Content, *duplicate.CreatedAt)
			})
			if !errors.Is(err, posts.ErrPostAlreadyExists) {
				t.Fatalf("error=%v, want post already exists", err)
			}

			got, err := repository.FindPostByID(t.Context(), original.Namespace, original.ID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, original) {
				t.Errorf("duplicate changed original: %+v", got)
			}
			var posts, revisions int
			if err := db.Pool.QueryRow(t.Context(), "select (select count(*) from posts), (select count(*) from posts_content)").Scan(&posts, &revisions); err != nil {
				t.Fatal(err)
			}
			if posts != 1 || revisions != 1 {
				t.Errorf("duplicate left %d posts and %d revisions, want one each", posts, revisions)
			}
		})
	}

	otherNamespace := *original
	otherNamespace.ID = uuid.New()
	otherNamespace.Namespace = "other"
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		contentID := uuid.New()
		if err := repository.CreatePost(ctx, &otherNamespace, contentID); err != nil {
			return err
		}

		return repository.CreatePostContent(ctx, otherNamespace.ID, contentID, otherNamespace.Title, otherNamespace.Content, *otherNamespace.CreatedAt)
	}); err != nil {
		t.Errorf("same slug in another namespace must be accepted: %v", err)
	}
}

func TestPostsRepositoryCreatePostRollsBackWhenRevisionFails(t *testing.T) {
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
	if _, err := db.Pool.Exec(t.Context(), "alter table posts_content add constraint reject_test_title check (title <> 'Rejected')"); err != nil {
		t.Fatal(err)
	}

	repository := posts.NewPostsRepository(db.Pool)
	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	item := &posts.Post{
		ID:        uuid.New(),
		Namespace: "main",
		Slug:      "failed-post",
		Title:     "Rejected",
		Content:   "Content",
		CreatedAt: &instant,
		UpdatedAt: &instant,
	}
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		contentID := uuid.New()
		if err := repository.CreatePost(ctx, item, contentID); err != nil {
			return err
		}

		return repository.CreatePostContent(ctx, item.ID, contentID, item.Title, item.Content, *item.CreatedAt)
	})
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != pgerrcode.CheckViolation || pgError.ConstraintName != "reject_test_title" {
		t.Fatalf("error=%v, want rejected content revision", err)
	}

	var posts, revisions int
	if err := db.Pool.QueryRow(t.Context(), "select (select count(*) from posts), (select count(*) from posts_content)").Scan(&posts, &revisions); err != nil {
		t.Fatal(err)
	}
	if posts != 0 || revisions != 0 {
		t.Errorf("failed creation left %d posts and %d revisions", posts, revisions)
	}
}

func TestPostsRepositoryDeletePost(t *testing.T) {
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
		insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'first-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '2026-09-11 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00'),
			('22222222-2222-4222-8222-222222222222', 'main', 'second-post', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00');
		insert into posts_content (id, post_id, title, content, created_at)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'Original title', 'Original content', '2026-09-10 12:00:00'),
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', 'Revised title', 'Revised content', '2026-09-11 12:00:00'),
			('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '22222222-2222-4222-8222-222222222222', 'Other post', 'Other content', '2026-09-10 13:00:00');`)
	if err != nil {
		t.Fatal(err)
	}

	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	deletedAt := time.Date(2026, 9, 12, 21, 0, 0, 0, time.FixedZone("UTC+9", 9*60*60))
	repository := posts.NewPostsRepository(db.Pool)
	const postSnapshotSQL = `select (to_jsonb(posts) - 'deleted_at')::text, deleted_at
		from posts where id = $1`
	const versionsSnapshotSQL = `select jsonb_agg(to_jsonb(posts_content) order by id)::text from posts_content`
	var before, versionsBefore string
	var initialDeletedAt *time.Time
	if err := db.Pool.QueryRow(t.Context(), postSnapshotSQL, id).Scan(&before, &initialDeletedAt); err != nil {
		t.Fatal(err)
	}
	if initialDeletedAt != nil {
		t.Fatal("seed post is already deleted")
	}
	if err := db.Pool.QueryRow(t.Context(), versionsSnapshotSQL).Scan(&versionsBefore); err != nil {
		t.Fatal(err)
	}

	if err := repository.DeletePost(t.Context(), "other", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	var wrongNamespaceDeletedAt *time.Time
	if err := db.Pool.QueryRow(t.Context(), "select deleted_at from posts where id = $1", id).Scan(&wrongNamespaceDeletedAt); err != nil {
		t.Fatal(err)
	}
	if wrongNamespaceDeletedAt != nil {
		t.Fatal("wrong namespace deleted the post")
	}
	if err := repository.DeletePost(t.Context(), "main", uuid.MustParse("99999999-9999-4999-8999-999999999999"), deletedAt); err != nil {
		t.Fatalf("missing post must be an idempotent success: %v", err)
	}
	if err := repository.DeletePost(t.Context(), "main", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.DeletePost(t.Context(), "main", id, deletedAt.Add(time.Hour)); err != nil {
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
		t.Errorf("deletion changed content versions:\nbefore %s\nafter %s", versionsBefore, versionsAfter)
	}
	var otherDeletedAt *time.Time
	if err := db.Pool.QueryRow(t.Context(), "select deleted_at from posts where id = $1", "22222222-2222-4222-8222-222222222222").Scan(&otherDeletedAt); err != nil {
		t.Fatalf("other post must remain present: %v", err)
	}
	if otherDeletedAt != nil {
		t.Errorf("other post was deleted at %v", otherDeletedAt)
	}
}

func TestPostsRepositoryFindUsesCurrentContentAndNamespace(t *testing.T) {
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
		insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'welcome', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '2026-09-11 10:00:00', '2026-09-10 08:00:00', '2026-09-11 09:30:00'),
			('22222222-2222-4222-8222-222222222222', 'other', 'welcome', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', null, '2026-09-10 08:00:00', '2026-09-11 09:30:00');
		insert into posts_content (id, post_id, title, content, created_at)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '11111111-1111-4111-8111-111111111111', 'Current title', 'Current content', '2026-09-11 09:30:00'),
			('cccccccc-cccc-4ccc-8ccc-cccccccccccc', '11111111-1111-4111-8111-111111111111', 'Old title', 'Old content', '2026-09-10 08:00:00'),
			('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', '22222222-2222-4222-8222-222222222222', 'Other title', 'Other content', '2026-09-11 09:30:00')`)
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
		"slug": func() (*posts.Post, error) { return repository.FindPostBySlug(t.Context(), "main", "welcome") },
		"ID":   func() (*posts.Post, error) { return repository.FindPostByID(t.Context(), "main", id) },
	} {
		got, err := find()
		if err != nil {
			t.Fatalf("%s lookup: %v", name, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s lookup=%+v, want %+v", name, got, want)
		}
	}
	other, err := repository.FindPostBySlug(t.Context(), "other", "welcome")
	if err != nil {
		t.Fatal(err)
	}
	if other.Title != "Other title" || other.PublishedAt != nil {
		t.Errorf("other namespace draft=%+v", other)
	}
	if _, err := repository.FindPostByID(t.Context(), "other", id); !errors.Is(err, posts.ErrPostNotFound) {
		t.Errorf("wrong namespace error=%v, want post not found", err)
	}

	wantRollback := errors.New("roll back post read")
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
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
	got, err := repository.FindPostByID(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rollback changed persisted post: %+v", got)
	}

	if _, err := db.Pool.Exec(t.Context(), "update posts set deleted_at = '2026-09-12 12:00:00' where id = $1", id); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindPostBySlug(t.Context(), "main", "welcome"); !errors.Is(err, posts.ErrPostNotFound) {
		t.Errorf("deleted slug error=%v, want post not found", err)
	}
	if _, err := repository.FindPostByID(t.Context(), "main", id); !errors.Is(err, posts.ErrPostNotFound) {
		t.Errorf("deleted ID error=%v, want post not found", err)
	}
}

func TestPostsRepositoryListPosts(t *testing.T) {
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
		insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
		values
			('10000000-0000-4000-8000-000000000001', 'main', 'older', '20000000-0000-4000-8000-000000000001', '2026-09-11 12:00:00', '2026-09-10 10:00:00', '2026-09-11 12:00:00', null),
			('10000000-0000-4000-8000-000000000002', 'main', 'boundary', '20000000-0000-4000-8000-000000000002', '2026-09-12 12:00:00', '2026-09-11 10:00:00', '2026-09-12 11:00:00', null),
			('10000000-0000-4000-8000-000000000003', 'main', 'scheduled', '20000000-0000-4000-8000-000000000003', '2026-09-12 12:00:01', '2026-09-11 11:00:00', '2026-09-12 11:00:00', null),
			('10000000-0000-4000-8000-000000000004', 'main', 'draft', '20000000-0000-4000-8000-000000000004', null, '2026-09-11 12:00:00', '2026-09-12 11:00:00', null),
			('10000000-0000-4000-8000-000000000005', 'main', 'deleted', '20000000-0000-4000-8000-000000000005', '2026-09-11 12:00:00', '2026-09-11 13:00:00', '2026-09-12 11:00:00', '2026-09-12 11:00:00'),
			('10000000-0000-4000-8000-000000000006', 'other', 'other', '20000000-0000-4000-8000-000000000006', '2026-09-11 12:00:00', '2026-09-11 14:00:00', '2026-09-12 11:00:00', null),
			('10000000-0000-4000-8000-000000000007', 'main', 'tied', '20000000-0000-4000-8000-000000000007', '2026-09-12 12:00:00', '2026-09-11 10:00:00', '2026-09-12 11:00:00', null);
		insert into posts_content (id, post_id, title, content, created_at)
		values
			('20000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000001', 'Older', 'Older content', '2026-09-10 10:00:00'),
			('20000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000002', 'Boundary', 'Boundary content', '2026-09-11 10:00:00'),
			('20000000-0000-4000-8000-000000000003', '10000000-0000-4000-8000-000000000003', 'Scheduled', 'Scheduled content', '2026-09-11 11:00:00'),
			('20000000-0000-4000-8000-000000000004', '10000000-0000-4000-8000-000000000004', 'Draft', 'Draft content', '2026-09-11 12:00:00'),
			('20000000-0000-4000-8000-000000000005', '10000000-0000-4000-8000-000000000005', 'Deleted', 'Deleted content', '2026-09-11 13:00:00'),
			('20000000-0000-4000-8000-000000000006', '10000000-0000-4000-8000-000000000006', 'Other', 'Other content', '2026-09-11 14:00:00'),
			('20000000-0000-4000-8000-000000000007', '10000000-0000-4000-8000-000000000007', 'Current title', 'Current content', '2026-09-12 11:00:00'),
			('20000000-0000-4000-8000-000000000008', '10000000-0000-4000-8000-000000000007', 'Old title', 'Old content', '2026-09-11 10:00:00');`)
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
			items, total, err := repository.ListPosts(t.Context(), test.namespace, test.includeDrafts, test.cutoff, test.limit, test.offset)
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

func TestPostsRepositoryUpdatePost(t *testing.T) {
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
		insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'original-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', null, '2026-09-10 12:00:00', '2026-09-10 12:00:00'),
			('22222222-2222-4222-8222-222222222222', 'main', 'other-post', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00');
		insert into posts_content (id, post_id, title, content, created_at)
		values
			('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '11111111-1111-4111-8111-111111111111', 'Original title', 'Original content', '2026-09-10 12:00:00'),
			('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', '22222222-2222-4222-8222-222222222222', 'Other title', 'Other content', '2026-09-10 13:00:00');
		alter table posts_content add constraint reject_update_title check (title <> 'rejected revision');`)
	if err != nil {
		t.Fatal(err)
	}

	repository := posts.NewPostsRepository(db.Pool)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	original, err := repository.FindPostByID(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}

	metadata := *original
	metadata.Slug = "renamed-post"
	publishedAt := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	metadata.PublishedAt = &publishedAt
	metadata.UpdatedAt = &publishedAt
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.UpdatePost(ctx, &metadata, nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := repository.FindPostByID(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, &metadata) {
		t.Errorf("metadata update=%+v, want %+v", got, metadata)
	}
	var contentID uuid.UUID
	var count int
	if err := db.Pool.QueryRow(t.Context(), `select current_content_id, (select count(*) from posts_content where post_id = $1) from posts where id = $1`, id).Scan(&contentID, &count); err != nil {
		t.Fatal(err)
	}
	if contentID != uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa") || count != 1 {
		t.Errorf("metadata update changed revision: current=%v, count=%d", contentID, count)
	}

	updated := metadata
	updated.Title = "Revised title"
	updated.Content = "Revised content"
	updated.PublishedAt = nil
	updatedAt := publishedAt.Add(time.Hour)
	updated.UpdatedAt = &updatedAt
	stop := errors.New("roll back revised post")
	for _, rollback := range []bool{true, false} {
		err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
			contentID := uuid.New()
			if err := repository.UpdatePost(ctx, &updated, &contentID); err != nil {
				return err
			}
			if err := repository.CreatePostContent(ctx, updated.ID, contentID, updated.Title, updated.Content, *updated.UpdatedAt); err != nil {
				return err
			}
			got, err := repository.FindPostByID(ctx, "main", id)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(got, &updated) {
				t.Errorf("read within transaction=%+v, want %+v", got, updated)
			}
			outside, err := repository.FindPostByID(t.Context(), "main", id)
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
		got, err := repository.FindPostByID(t.Context(), "main", id)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("rollback=%t: persisted post=%+v, want %+v", rollback, got, want)
		}
		if err := db.Pool.QueryRow(t.Context(), "select count(*) from posts_content where post_id = $1", id).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != wantCount {
			t.Errorf("rollback=%t: revision count=%d, want %d", rollback, count, wantCount)
		}
	}

	var title, body string
	var createdAt time.Time
	if err := db.Pool.QueryRow(t.Context(), `select title, content, created_at from posts_content where id = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa'`).Scan(&title, &body, &createdAt); err != nil {
		t.Fatal(err)
	}
	if title != original.Title || body != original.Content || !createdAt.Equal(*original.CreatedAt) {
		t.Errorf("original revision changed: title=%q, content=%q, created_at=%v", title, body, createdAt)
	}
	if err := db.Pool.QueryRow(t.Context(), `select posts_content.created_at from posts join posts_content on posts_content.id = posts.current_content_id where posts.id = $1`, id).Scan(&createdAt); err != nil {
		t.Fatal(err)
	}
	if !createdAt.Equal(*updated.UpdatedAt) {
		t.Errorf("revision timestamp=%v, want explicit update time %v", createdAt, updated.UpdatedAt)
	}

	const snapshotSQL = `select jsonb_build_object(
		'posts', (select jsonb_agg(to_jsonb(p) order by id) from posts p),
		'content', (select jsonb_agg(to_jsonb(c) order by id) from posts_content c)
	)::text`
	for _, test := range []struct {
		name   string
		change func(*posts.Post)
		want   error
	}{
		{name: "wrong namespace", change: func(post *posts.Post) { post.Namespace = "other" }, want: posts.ErrPostNotFound},
		{name: "missing post", change: func(post *posts.Post) { post.ID = uuid.MustParse("99999999-9999-4999-8999-999999999999") }, want: posts.ErrPostNotFound},
		{name: "duplicate slug", change: func(post *posts.Post) { post.Slug = "other-post" }, want: posts.ErrPostAlreadyExists},
		{name: "revision insertion failure", change: func(post *posts.Post) {
			post.Slug = "rolled-back-post"
			post.Title = "rejected revision"
			updatedAt := post.UpdatedAt.Add(time.Hour)
			post.UpdatedAt = &updatedAt
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
				if err := repository.UpdatePost(ctx, &attempt, &contentID); err != nil {
					return err
				}

				return repository.CreatePostContent(ctx, attempt.ID, contentID, attempt.Title, attempt.Content, *attempt.UpdatedAt)
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

	if _, err := db.Pool.Exec(t.Context(), "update posts set deleted_at = $1 where id = $2", *updated.UpdatedAt, id); err != nil {
		t.Fatal(err)
	}
	var before, after string
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&before); err != nil {
		t.Fatal(err)
	}
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		contentID := uuid.New()
		if err := repository.UpdatePost(ctx, &updated, &contentID); err != nil {
			return err
		}

		return repository.CreatePostContent(ctx, updated.ID, contentID, updated.Title, updated.Content, *updated.UpdatedAt)
	})
	if !errors.Is(err, posts.ErrPostNotFound) {
		t.Errorf("deleted post error=%v, want not found", err)
	}
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("deleted post update changed storage:\nbefore %s\nafter %s", before, after)
	}
}

func TestPostsRepositoryGetPostVersion(t *testing.T) {
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
		insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'published-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '2026-09-11 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00', null),
			('22222222-2222-4222-8222-222222222222', 'main', 'draft-post', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00', null),
			('33333333-3333-4333-8333-333333333333', 'main', 'deleted-post', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00', '2026-09-11 12:00:00');
		insert into posts_content (id, post_id, title, content, created_at)
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

	repository := posts.NewPostsRepository(db.Pool)
	postID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
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
			got, err := repository.GetPostVersion(t.Context(), "main", postID, id)
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != id || got.Version != test.version || got.Title != test.title || got.Content != test.body || !got.CreatedAt.Equal(test.createdAt) {
				t.Errorf("version=%+v, want ID=%s ordinal=%d title=%q content=%q created_at=%v", got, id, test.version, test.title, test.body, test.createdAt)
			}
		})
	}

	for _, test := range []struct {
		name      string
		namespace string
		postID    string
		contentID string
	}{
		{"wrong namespace", "other", postID.String(), "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"},
		{"wrong post", "main", "22222222-2222-4222-8222-222222222222", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1"},
		{"missing revision", "main", postID.String(), "99999999-9999-4999-8999-999999999999"},
		{"deleted post", "main", "33333333-3333-4333-8333-333333333333", "cccccccc-cccc-4ccc-8ccc-ccccccccccc1"},
		{"missing parent", "main", "44444444-4444-4444-8444-444444444444", "dddddddd-dddd-4ddd-8ddd-ddddddddddd1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := repository.GetPostVersion(t.Context(), test.namespace, uuid.MustParse(test.postID), uuid.MustParse(test.contentID))
			if !errors.Is(err, posts.ErrPostNotFound) {
				t.Errorf("error=%v, want post not found", err)
			}
		})
	}
}

func TestPostsRepositoryListPostVersions(t *testing.T) {
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
		insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
		values
			('11111111-1111-4111-8111-111111111111', 'main', 'scheduled', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '2026-09-14 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00', null),
			('22222222-2222-4222-8222-222222222222', 'main', 'deleted', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', null, '2026-09-10 12:00:00', '2026-09-11 12:00:00', '2026-09-12 12:00:00');
		insert into posts_content (id, post_id, title, content, created_at)
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
	want := []posts.PostVersion{
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
	repository := posts.NewPostsRepository(db.Pool)
	versions, err := repository.ListPostVersions(t.Context(), "main", id)
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
			versions, err := repository.ListPostVersions(t.Context(), test.namespace, test.id)
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
		if _, err := executor.Exec(ctx, "update posts_content set title = 'Uncommitted' where id = $1", want[0].ID); err != nil {
			return err
		}

		versions, err := repository.ListPostVersions(ctx, "main", id)
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
	versions, err = repository.ListPostVersions(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(versions, want) {
		t.Errorf("rolled-back history=%+v, want %+v", versions, want)
	}
}
