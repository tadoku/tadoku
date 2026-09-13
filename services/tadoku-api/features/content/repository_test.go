package content_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestRepositoryUsesSuppliedPolicyAndTransaction(t *testing.T) {
	t.Parallel()
	db := testpostgres.New(t)
	cutoff := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	id := db.SeedAnnouncement(t, "main", "older", cutoff.Add(-time.Hour), cutoff.Add(time.Hour), false)
	db.SeedAnnouncement(t, "main", "newer", cutoff, cutoff.Add(time.Hour), false)
	repository := content.NewRepository(db.Pool)
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
}

func TestEmptyNamespaceIsRejectedBeforeStorage(t *testing.T) {
	service := content.NewService(nil)
	_, err := service.ActiveAnnouncements(context.Background(), "")
	if !errors.Is(err, content.ErrInvalidNamespace) {
		t.Errorf("error=%v want invalid namespace", err)
	}
}
