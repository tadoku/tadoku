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

func TestUpdateAnnouncementParametersValidation(t *testing.T) {
	parameters := content.UpdateAnnouncementParameters{
		Title:    "Title",
		Content:  "Content",
		Style:    "info",
		StartsAt: time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
	}
	if err := parameters.Validate(""); !errors.Is(err, content.ErrInvalidAnnouncement) {
		t.Errorf("error=%v, want invalid announcement", err)
	}
}

func TestAnnouncementUpdateRepositoryTransaction(t *testing.T) {
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
	repository := content.NewAnnouncementsRepository(db.Pool)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	original, err := repository.FindAnnouncementByID(t.Context(), "main", id)
	if err != nil {
		t.Fatal(err)
	}

	href := "https://example.test/updated"
	updated := *original
	updated.Title = "Updated"
	updated.Content = "Updated content"
	updated.Style = "success"
	updated.Href = &href
	updated.StartsAt = original.StartsAt.Add(time.Hour)
	updated.EndsAt = original.EndsAt.Add(time.Hour)
	updated.UpdatedAt = original.UpdatedAt.Add(time.Hour)
	stop := errors.New("roll back updated announcement")
	for _, rollback := range []bool{true, false} {
		err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
			if err := repository.UpdateAnnouncement(ctx, &updated); err != nil {
				return err
			}
			got, err := repository.FindAnnouncementByID(ctx, "main", id)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(got, &updated) {
				t.Errorf("read within transaction=%+v, want %+v", got, updated)
			}
			outside, err := repository.FindAnnouncementByID(t.Context(), "main", id)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(outside, original) {
				t.Errorf("uncommitted update escaped transaction: %+v", outside)
			}
			if rollback {
				return stop
			}
			return nil
		})
		want := &updated
		if rollback {
			want = original
			if !errors.Is(err, stop) {
				t.Fatalf("rollback error=%v, want %v", err, stop)
			}
		} else if err != nil {
			t.Fatal(err)
		}
		got, err := repository.FindAnnouncementByID(t.Context(), "main", id)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("rollback=%t: read after transaction=%+v, want %+v", rollback, got, want)
		}
	}

	updated.Namespace = "other"
	if err := repository.UpdateAnnouncement(t.Context(), &updated); !errors.Is(err, content.ErrAnnouncementNotFound) {
		t.Errorf("wrong namespace error=%v, want not found", err)
	}
	updated.Namespace = "main"
	if _, err := db.Pool.Exec(t.Context(), "update announcements set deleted_at = $1 where id = $2", updated.UpdatedAt, id); err != nil {
		t.Fatal(err)
	}
	if err := repository.UpdateAnnouncement(t.Context(), &updated); !errors.Is(err, content.ErrAnnouncementNotFound) {
		t.Errorf("deleted announcement error=%v, want not found", err)
	}
}
