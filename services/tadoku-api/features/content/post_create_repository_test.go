package content_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
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
	repository := content.NewPostsRepository(db.Pool)
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
			item := &content.Post{
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
				if err := repository.CreatePost(ctx, item); err != nil {
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

	repository := content.NewPostsRepository(db.Pool)
	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	original := &content.Post{
		ID:        uuid.New(),
		Namespace: "main",
		Slug:      "first-post",
		Title:     "Original title",
		Content:   "Original content",
		CreatedAt: &instant,
		UpdatedAt: &instant,
	}
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.CreatePost(ctx, original)
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
				return repository.CreatePost(ctx, &duplicate)
			})
			if !errors.Is(err, content.ErrPostAlreadyExists) {
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
		return repository.CreatePost(ctx, &otherNamespace)
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

	repository := content.NewPostsRepository(db.Pool)
	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	item := &content.Post{
		ID:        uuid.New(),
		Namespace: "main",
		Slug:      "failed-post",
		Title:     "Rejected",
		Content:   "Content",
		CreatedAt: &instant,
		UpdatedAt: &instant,
	}
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.CreatePost(ctx, item)
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
