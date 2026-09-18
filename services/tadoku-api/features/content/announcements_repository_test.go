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

func TestAnnouncementsRepositoryListAnnouncementsBreaksCreatedAtTiesByID(t *testing.T) {
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
	if err := db.Reset(t.Context()); err != nil {
		t.Fatal(err)
	}

	_, err = db.Pool.Exec(t.Context(), `
		insert into announcements (id, namespace, title, content, starts_at, ends_at, created_at, updated_at)
		values
			('10000000-0000-4000-8000-000000000001', 'main', 'one', 'one', '2026-09-12 10:00:00', '2026-09-12 14:00:00', '2026-09-12 11:00:00', '2026-09-12 11:00:00'),
			('10000000-0000-4000-8000-000000000002', 'main', 'two', 'two', '2026-09-12 10:00:00', '2026-09-12 14:00:00', '2026-09-12 11:00:00', '2026-09-12 11:00:00'),
			('10000000-0000-4000-8000-000000000003', 'main', 'three', 'three', '2026-09-12 10:00:00', '2026-09-12 14:00:00', '2026-09-12 11:00:00', '2026-09-12 11:00:00')`)
	if err != nil {
		t.Fatal(err)
	}

	want := []uuid.UUID{
		uuid.MustParse("10000000-0000-4000-8000-000000000003"),
		uuid.MustParse("10000000-0000-4000-8000-000000000002"),
		uuid.MustParse("10000000-0000-4000-8000-000000000001"),
	}
	repository := content.NewAnnouncementsRepository(db.Pool)
	for attempt := 0; attempt < 5; attempt++ {
		items, err := repository.ListAnnouncements(t.Context(), "main", 3, 0)
		if err != nil {
			t.Fatal(err)
		}

		got := make([]uuid.UUID, 0, len(items))
		for _, item := range items {
			got = append(got, item.ID)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("attempt %d IDs=%v, want %v", attempt+1, got, want)
		}
	}
}

func TestAnnouncementsRepositoryDeleteAnnouncement(t *testing.T) {
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
	if err := db.Reset(t.Context(), "testdata/announcements.sql"); err != nil {
		t.Fatal(err)
	}

	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	deletedAt := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	repository := content.NewAnnouncementsRepository(db.Pool)
	const snapshotSQL = `select (to_jsonb(announcements) - 'deleted_at')::text, deleted_at
		from announcements where id = $1`
	var before string
	var initialDeletedAt *time.Time
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL, id).Scan(&before, &initialDeletedAt); err != nil {
		t.Fatal(err)
	}
	if initialDeletedAt != nil {
		t.Fatal("seed announcement is already deleted")
	}

	if err := repository.DeleteAnnouncement(t.Context(), "other", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindAnnouncementByID(t.Context(), "main", id); err != nil {
		t.Fatalf("wrong namespace must not delete the announcement: %v", err)
	}
	if err := repository.DeleteAnnouncement(t.Context(), "main", id, deletedAt); err != nil {
		t.Fatal(err)
	}
	if err := repository.DeleteAnnouncement(t.Context(), "main", id, deletedAt.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	var after string
	var gotDeletedAt time.Time
	if err := db.Pool.QueryRow(t.Context(), snapshotSQL, id).Scan(&after, &gotDeletedAt); err != nil {
		t.Fatalf("soft-deleted row must still exist: %v", err)
	}
	if after != before {
		t.Errorf("deletion changed fields other than deleted_at:\nbefore %s\nafter %s", before, after)
	}
	if !gotDeletedAt.Equal(deletedAt) {
		t.Errorf("deleted_at=%v, want original timestamp %v", gotDeletedAt, deletedAt)
	}
	otherID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	if _, err := repository.FindAnnouncementByID(t.Context(), "main", otherID); err != nil {
		t.Errorf("other announcements must remain visible: %v", err)
	}
}

func TestAnnouncementsRepositoryCreateAnnouncement(t *testing.T) {
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
	repository := content.NewAnnouncementsRepository(db.Pool)
	instant := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	emptyHref, href := "", "https://example.test/announcement"
	for _, test := range []struct {
		name string
		href *string
	}{
		{name: "null href"},
		{name: "empty href", href: &emptyHref},
		{name: "href", href: &href},
	} {
		t.Run(test.name, func(t *testing.T) {
			item := &content.Announcement{
				ID:        uuid.New(),
				Namespace: "main",
				Title:     "Title",
				Content:   "Content",
				Style:     "warning",
				Href:      test.href,
				StartsAt:  instant,
				EndsAt:    instant.Add(time.Hour),
				CreatedAt: instant.Add(-2 * time.Hour),
				UpdatedAt: instant.Add(-time.Hour),
			}
			if err := repository.CreateAnnouncement(t.Context(), item); err != nil {
				t.Fatal(err)
			}
			got, err := repository.FindAnnouncementByID(t.Context(), item.Namespace, item.ID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, item) {
				t.Errorf("persisted=%+v, want %+v", got, item)
			}
			duplicate := *item
			duplicate.Title = "must not replace the original"
			if err := repository.CreateAnnouncement(t.Context(), &duplicate); err == nil {
				t.Fatal("duplicate ID was accepted")
			}
			got, err = repository.FindAnnouncementByID(t.Context(), item.Namespace, item.ID)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, item) {
				t.Errorf("duplicate creation changed original: %+v", got)
			}
		})
	}
}

func TestAnnouncementsRepositoryPreservesTimestampInstants(t *testing.T) {
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

	location := time.FixedZone("UTC+9", 9*60*60)
	startsAt := time.Date(2026, 9, 13, 10, 0, 0, 0, location)
	item := &content.Announcement{
		ID:        uuid.New(),
		Namespace: "main",
		Title:     "Title",
		Content:   "Content",
		Style:     "info",
		StartsAt:  startsAt,
		EndsAt:    startsAt.Add(time.Hour),
		CreatedAt: startsAt.Add(-2 * time.Hour),
		UpdatedAt: startsAt.Add(-time.Hour),
	}
	repository := content.NewAnnouncementsRepository(db.Pool)
	if err := repository.CreateAnnouncement(t.Context(), item); err != nil {
		t.Fatal(err)
	}

	got, err := repository.FindAnnouncementByID(t.Context(), item.Namespace, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	for name, timestamps := range map[string][2]time.Time{
		"starts_at":  {got.StartsAt, item.StartsAt},
		"ends_at":    {got.EndsAt, item.EndsAt},
		"created_at": {got.CreatedAt, item.CreatedAt},
		"updated_at": {got.UpdatedAt, item.UpdatedAt},
	} {
		if !timestamps[0].Equal(timestamps[1]) {
			t.Errorf("%s instant=%v, want %v", name, timestamps[0], timestamps[1])
		}
	}
}

func TestEmptyNamespaceIsRejectedBeforeStorage(t *testing.T) {
	service := content.NewService(nil)
	_, err := service.ListActiveAnnouncements(context.Background(), "")
	if !errors.Is(err, content.ErrInvalidNamespace) {
		t.Errorf("error=%v want invalid namespace", err)
	}
}
