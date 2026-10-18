package posts_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestPostsRepositoryCreatePost(t *testing.T) {
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

			err := postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
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

			got, err := repository.FindPostByID(tenantCtx, item.Namespace, item.ID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, &want) {
				t.Errorf("committed post=%+v, want %+v", got, want)
			}

			var contentID uuid.UUID
			var contentCreatedAt time.Time
			if err := db.Pool.QueryRow(tenantCtx, `
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
			if err := db.Pool.QueryRow(tenantCtx, "select count(*) from posts_content where post_id = $1", item.ID).Scan(&revisions); err != nil {
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

	if err := postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
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

			err := postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
				contentID := uuid.New()
				if err := repository.CreatePost(ctx, &duplicate, contentID); err != nil {
					return err
				}

				return repository.CreatePostContent(ctx, duplicate.ID, contentID, duplicate.Title, duplicate.Content, *duplicate.CreatedAt)
			})
			if !errors.Is(err, posts.ErrPostAlreadyExists) {
				t.Fatalf("error=%v, want post already exists", err)
			}

			got, err := repository.FindPostByID(tenantCtx, original.Namespace, original.ID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, original) {
				t.Errorf("duplicate changed original: %+v", got)
			}

			var posts, revisions int
			if err := db.Pool.QueryRow(tenantCtx, "select (select count(*) from posts), (select count(*) from posts_content)").Scan(&posts, &revisions); err != nil {
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
	if err := postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
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
	if _, err := db.Pool.Exec(tenantCtx, "alter table posts_content add constraint reject_test_title check (title <> 'Rejected')"); err != nil {
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

	err = postgres.RunInTransaction(tenantCtx, db.Pool, func(ctx context.Context) error {
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
	if err := db.Pool.QueryRow(tenantCtx, "select (select count(*) from posts), (select count(*) from posts_content)").Scan(&posts, &revisions); err != nil {
		t.Fatal(err)
	}
	if posts != 0 || revisions != 0 {
		t.Errorf("failed creation left %d posts and %d revisions", posts, revisions)
	}
}

func TestPostsRepositoryDeletePost(t *testing.T) {
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
			 'main', 'first-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
			 '2026-09-11 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00'),
			('tadoku/prod', '22222222-2222-4222-8222-222222222222',
			 'main', 'second-post', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1',
			 null, '2026-09-10 13:00:00', '2026-09-10 13:00:00');
		insert into posts_content (tenant, id, post_id, title, content, created_at)
		values
			('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111',
			 'Original title', 'Original content', '2026-09-10 12:00:00'),
			('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111',
			 'Revised title', 'Revised content', '2026-09-11 12:00:00'),
			('tadoku/prod', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '22222222-2222-4222-8222-222222222222',
			 'Other post', 'Other content', '2026-09-10 13:00:00');`)
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
	if err := db.Pool.QueryRow(tenantCtx, postSnapshotSQL, id).Scan(&before, &initialDeletedAt); err != nil {
		t.Fatal(err)
	}
	if initialDeletedAt != nil {
		t.Fatal("seed post is already deleted")
	}
	if err := db.Pool.QueryRow(tenantCtx, versionsSnapshotSQL).Scan(&versionsBefore); err != nil {
		t.Fatal(err)
	}

	if err := repository.DeletePost(tenantCtx, "other", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	var wrongNamespaceDeletedAt *time.Time
	if err := db.Pool.QueryRow(tenantCtx, "select deleted_at from posts where id = $1", id).Scan(&wrongNamespaceDeletedAt); err != nil {
		t.Fatal(err)
	}
	if wrongNamespaceDeletedAt != nil {
		t.Fatal("wrong namespace deleted the post")
	}

	if err := repository.DeletePost(tenantCtx, "main", uuid.MustParse("99999999-9999-4999-8999-999999999999"), deletedAt); err != nil {
		t.Fatalf("missing post must be an idempotent success: %v", err)
	}

	if err := repository.DeletePost(tenantCtx, "main", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.DeletePost(tenantCtx, "main", id, deletedAt.Add(time.Hour)); err != nil {
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
		t.Errorf("deletion changed content versions:\nbefore %s\nafter %s", versionsBefore, versionsAfter)
	}

	var otherDeletedAt *time.Time
	if err := db.Pool.QueryRow(tenantCtx, "select deleted_at from posts where id = $1", "22222222-2222-4222-8222-222222222222").Scan(&otherDeletedAt); err != nil {
		t.Fatalf("other post must remain present: %v", err)
	}
	if otherDeletedAt != nil {
		t.Errorf("other post was deleted at %v", otherDeletedAt)
	}
}
