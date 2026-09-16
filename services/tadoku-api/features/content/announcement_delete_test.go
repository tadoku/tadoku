package content_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestDeleteAnnouncementUsesBusinessTimeAndTransaction(t *testing.T) {
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.Reset(t.Context(), "testdata/announcements.sql"); err != nil {
		t.Fatal(err)
	}

	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	deletedAt := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	repository := content.NewAnnouncementsRepository(db.Pool)
	service := content.NewService(repository)
	wantRollback := errors.New("stop after delete")
	timex.TheWorld(deletedAt, func() {
		err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
			if err := service.DeleteAnnouncement(ctx, "main", id); err != nil {
				return err
			}
			if _, err := repository.FindAnnouncementByID(ctx, "main", id); !errors.Is(err, content.ErrAnnouncementNotFound) {
				t.Errorf("deleted announcement lookup error=%v, want not found", err)
			}
			executor, err := postgres.Executor(ctx, db.Pool)
			if err != nil {
				return err
			}
			var got time.Time
			if err := executor.QueryRow(ctx, "select deleted_at from announcements where id = $1", id).Scan(&got); err != nil {
				return err
			}
			if !got.Equal(deletedAt) {
				t.Errorf("deleted_at=%v, want business time %v", got, deletedAt)
			}
			return wantRollback
		})
	})
	if !errors.Is(err, wantRollback) {
		t.Fatalf("transaction error=%v, want rollback", err)
	}
	if _, err := repository.FindAnnouncementByID(t.Context(), "main", id); err != nil {
		t.Errorf("announcement should survive rollback: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := repository.DeleteAnnouncement(ctx, "main", id, deletedAt); !errors.Is(err, context.Canceled) {
		t.Errorf("canceled delete error=%v, want context canceled", err)
	}
	if _, err := repository.FindAnnouncementByID(t.Context(), "main", id); err != nil {
		t.Errorf("announcement should survive canceled delete: %v", err)
	}

	if err := repository.DeleteAnnouncement(t.Context(), "main", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.DeleteAnnouncement(t.Context(), "main", id, deletedAt.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	var got time.Time
	if err := db.Pool.QueryRow(t.Context(), "select deleted_at from announcements where id = $1", id).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if !got.Equal(deletedAt) {
		t.Errorf("repeated delete changed deleted_at to %v, want %v", got, deletedAt)
	}
}
