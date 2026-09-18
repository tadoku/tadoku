package content_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

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

	repository := content.NewPostsRepository(db.Pool)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	publishedAt := time.Date(2026, 9, 11, 10, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 9, 11, 9, 30, 0, 0, time.UTC)
	want := &content.Post{
		ID:          id,
		Namespace:   "main",
		Slug:        "welcome",
		Title:       "Current title",
		Content:     "Current content",
		PublishedAt: &publishedAt,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}
	for name, find := range map[string]func() (*content.Post, error){
		"slug": func() (*content.Post, error) { return repository.FindPostBySlug(t.Context(), "main", "welcome") },
		"ID":   func() (*content.Post, error) { return repository.FindPostByID(t.Context(), "main", id) },
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
	if _, err := repository.FindPostByID(t.Context(), "other", id); !errors.Is(err, content.ErrPostNotFound) {
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
	if _, err := repository.FindPostBySlug(t.Context(), "main", "welcome"); !errors.Is(err, content.ErrPostNotFound) {
		t.Errorf("deleted slug error=%v, want post not found", err)
	}
	if _, err := repository.FindPostByID(t.Context(), "main", id); !errors.Is(err, content.ErrPostNotFound) {
		t.Errorf("deleted ID error=%v, want post not found", err)
	}
}
