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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestCreateAnnouncementValidationBeforeStorage(t *testing.T) {
	valid := content.CreateAnnouncementRequest{
		ID:        uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Namespace: "main",
		Title:     "Title",
		Content:   "Content",
		Style:     "info",
		StartsAt:  time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
		EndsAt:    time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
	}
	for _, test := range []struct {
		name   string
		change func(*content.CreateAnnouncementRequest)
	}{
		{name: "ID", change: func(r *content.CreateAnnouncementRequest) { r.ID = uuid.Nil }},
		{name: "namespace", change: func(r *content.CreateAnnouncementRequest) { r.Namespace = "" }},
		{name: "title", change: func(r *content.CreateAnnouncementRequest) { r.Title = "" }},
		{name: "content", change: func(r *content.CreateAnnouncementRequest) { r.Content = "" }},
		{name: "style", change: func(r *content.CreateAnnouncementRequest) { r.Style = "other" }},
		{name: "starts at", change: func(r *content.CreateAnnouncementRequest) { r.StartsAt = time.Time{} }},
		{name: "ends at", change: func(r *content.CreateAnnouncementRequest) { r.EndsAt = time.Time{} }},
		{name: "equal dates", change: func(r *content.CreateAnnouncementRequest) { r.EndsAt = r.StartsAt }},
		{name: "reversed dates", change: func(r *content.CreateAnnouncementRequest) { r.EndsAt = r.StartsAt.Add(-time.Second) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := valid
			test.change(&request)
			_, err := content.NewService(nil).CreateAnnouncement(t.Context(), request)
			if !errors.Is(err, content.ErrInvalidAnnouncement) {
				t.Errorf("error=%v, want invalid announcement", err)
			}
		})
	}
}

func TestCreateAnnouncementBusinessTimeAndRepositoryTransactions(t *testing.T) {
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
	service := content.NewService(repository)
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	emptyHref := ""
	timex.TheWorld(now, func() {
		for _, style := range []string{"info", "success", "warning", "error"} {
			t.Run(style, func(t *testing.T) {
				request := content.CreateAnnouncementRequest{
					ID:        uuid.New(),
					Namespace: "main",
					Title:     " ", // Legacy required validation does not trim strings.
					Content:   "Content",
					Style:     style,
					Href:      &emptyHref,
					StartsAt:  now.Add(-time.Hour),
					EndsAt:    now.Add(time.Hour),
				}
				created, err := service.CreateAnnouncement(t.Context(), request)
				if err != nil {
					t.Fatal(err)
				}
				if !created.CreatedAt.Equal(now) || !created.UpdatedAt.Equal(now) {
					t.Errorf("business timestamps=%s, %s, want %s", created.CreatedAt, created.UpdatedAt, now)
				}
				persisted, err := repository.FindAnnouncementByID(t.Context(), "main", request.ID)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(created, persisted) {
					t.Errorf("persisted=%+v, want %+v", persisted, created)
				}
			})
		}
	})

	item := &content.Announcement{
		ID:        uuid.New(),
		Namespace: "main",
		Title:     "Transactional",
		Content:   "Content",
		Style:     "info",
		StartsAt:  now,
		EndsAt:    now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	wantRollback := errors.New("rollback test transaction")
	err = postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		if err := repository.CreateAnnouncement(ctx, item); err != nil {
			return err
		}
		persisted, err := repository.FindAnnouncementByID(ctx, "main", item.ID)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(item, persisted) {
			t.Errorf("transactional read=%+v, want %+v", persisted, item)
		}
		return wantRollback
	})
	if !errors.Is(err, wantRollback) {
		t.Fatalf("transaction error=%v, want rollback marker", err)
	}
	if _, err := repository.FindAnnouncementByID(t.Context(), "main", item.ID); !errors.Is(err, content.ErrAnnouncementNotFound) {
		t.Errorf("rolled back create error=%v, want not found", err)
	}
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.CreateAnnouncement(ctx, item)
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindAnnouncementByID(t.Context(), "main", item.ID); err != nil {
		t.Errorf("committed create: %v", err)
	}
}
