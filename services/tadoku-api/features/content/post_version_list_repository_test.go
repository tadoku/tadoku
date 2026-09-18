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
	want := []content.PostVersion{
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
	repository := content.NewPostsRepository(db.Pool)
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
