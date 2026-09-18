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

	repository := content.NewPostsRepository(db.Pool)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	original, err := repository.FindPostByID(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}

	metadata := *original
	metadata.Slug = "renamed-post"
	publishedAt := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)
	metadata.PublishedAt = &publishedAt
	metadata.UpdatedAt = publishedAt
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.UpdatePost(ctx, &metadata, false)
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
	updated.UpdatedAt = publishedAt.Add(time.Hour)
	stop := errors.New("roll back revised post")
	for _, rollback := range []bool{true, false} {
		err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
			if err := repository.UpdatePost(ctx, &updated, true); err != nil {
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
	if title != original.Title || body != original.Content || !createdAt.Equal(original.CreatedAt) {
		t.Errorf("original revision changed: title=%q, content=%q, created_at=%v", title, body, createdAt)
	}
	if err := db.Pool.QueryRow(t.Context(), `select posts_content.created_at from posts join posts_content on posts_content.id = posts.current_content_id where posts.id = $1`, id).Scan(&createdAt); err != nil {
		t.Fatal(err)
	}
	if !createdAt.Equal(updated.UpdatedAt) {
		t.Errorf("revision timestamp=%v, want explicit update time %v", createdAt, updated.UpdatedAt)
	}

	const snapshotSQL = `select jsonb_build_object(
		'posts', (select jsonb_agg(to_jsonb(p) order by id) from posts p),
		'content', (select jsonb_agg(to_jsonb(c) order by id) from posts_content c)
	)::text`
	for _, test := range []struct {
		name   string
		change func(*content.Post)
		want   error
	}{
		{name: "wrong namespace", change: func(post *content.Post) { post.Namespace = "other" }, want: content.ErrPostNotFound},
		{name: "missing post", change: func(post *content.Post) { post.ID = uuid.MustParse("99999999-9999-4999-8999-999999999999") }, want: content.ErrPostNotFound},
		{name: "duplicate slug", change: func(post *content.Post) { post.Slug = "other-post" }, want: content.ErrPostAlreadyExists},
		{name: "revision insertion failure", change: func(post *content.Post) {
			post.Slug = "rolled-back-post"
			post.Title = "rejected revision"
			post.UpdatedAt = post.UpdatedAt.Add(time.Hour)
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
				return repository.UpdatePost(ctx, &attempt, true)
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

	if _, err := db.Pool.Exec(t.Context(), "update posts set deleted_at = $1 where id = $2", updated.UpdatedAt, id); err != nil {
		t.Fatal(err)
	}
	var before, after string
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&before); err != nil {
		t.Fatal(err)
	}
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.UpdatePost(ctx, &updated, true)
	})
	if !errors.Is(err, content.ErrPostNotFound) {
		t.Errorf("deleted post error=%v, want not found", err)
	}
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("deleted post update changed storage:\nbefore %s\nafter %s", before, after)
	}
}
