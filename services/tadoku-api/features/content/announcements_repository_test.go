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
)

func TestAnnouncementsRepositoryUsesSuppliedPolicyAndTransaction(t *testing.T) {
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
	cutoff := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	id := "11111111-1111-4111-8111-111111111111"
	if err := db.Reset(t.Context(), "testdata/announcements.sql"); err != nil {
		t.Fatal(err)
	}
	repository := content.NewAnnouncementsRepository(db.Pool)
	items, err := repository.ListActiveAnnouncements(context.Background(), "main", cutoff, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "newer" {
		t.Fatalf("explicit limit/order: %+v", items)
	}
	wantRollback := errors.New("stop after read")
	err = postgres.RunInTransaction(context.Background(), db.Pool, func(ctx context.Context) error {
		executor, err := postgres.Executor(ctx, db.Pool)
		if err != nil {
			return err
		}
		if _, err := executor.Exec(ctx, "update announcements set title = 'uncommitted' where id = $1", id); err != nil {
			return err
		}
		items, err := repository.ListActiveAnnouncements(ctx, "main", cutoff.Add(-time.Minute), 2)
		if err != nil {
			return err
		}
		if len(items) != 1 || items[0].Title != "uncommitted" {
			t.Errorf("transaction read: %+v", items)
		}
		item, err := repository.FindAnnouncementByID(ctx, "main", uuid.MustParse(id))
		if err != nil {
			return err
		}
		if item.Title != "uncommitted" {
			t.Errorf("transaction lookup: %+v", item)
		}
		return wantRollback
	})
	if !errors.Is(err, wantRollback) {
		t.Fatalf("transaction error=%v", err)
	}
	items, err = repository.ListActiveAnnouncements(context.Background(), "main", cutoff.Add(-time.Minute), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "older" {
		t.Errorf("rolled back read: %+v", items)
	}
	item, err := repository.FindAnnouncementByID(t.Context(), "main", uuid.MustParse(id))
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "older" {
		t.Errorf("rolled back lookup: %+v", item)
	}
	_, err = repository.FindAnnouncementByID(t.Context(), "other", uuid.MustParse(id))
	if !errors.Is(err, content.ErrAnnouncementNotFound) {
		t.Errorf("wrong namespace error=%v, want announcement not found", err)
	}
}

func TestEmptyNamespaceIsRejectedBeforeStorage(t *testing.T) {
	service := content.NewService(nil)
	_, err := service.ListActiveAnnouncements(context.Background(), "")
	if !errors.Is(err, content.ErrInvalidNamespace) {
		t.Errorf("error=%v want invalid namespace", err)
	}
}
