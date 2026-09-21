package profile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestRepositorySynchronizesAndLocksLocalUsers(t *testing.T) {
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

	repository := NewRepository(db.Pool)
	userID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	sessionCreatedAt := time.Date(2026, time.September, 12, 10, 0, 0, 0, time.UTC)
	user := SignedUser{ID: userID, DisplayName: "First name", sessionCreatedAt: sessionCreatedAt}
	if err := repository.SynchronizeUser(t.Context(), user, sessionCreatedAt); err != nil {
		t.Fatal(err)
	}

	user.DisplayName = "Same-session name"
	if err := repository.SynchronizeUser(t.Context(), user, sessionCreatedAt.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	var displayName string
	var updatedAt time.Time
	if err := db.Pool.QueryRow(t.Context(), `select display_name, updated_at from users where id = $1`, userID).Scan(&displayName, &updatedAt); err != nil {
		t.Fatal(err)
	}
	if displayName != "First name" || !updatedAt.Equal(sessionCreatedAt) {
		t.Errorf("equal-session synchronization=(%q, %s), want original values", displayName, updatedAt)
	}

	user.DisplayName = "New session name"
	user.sessionCreatedAt = sessionCreatedAt.Add(time.Second)
	now := sessionCreatedAt.Add(2 * time.Hour)
	if err := repository.SynchronizeUser(t.Context(), user, now); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(t.Context(), `select display_name, updated_at from users where id = $1`, userID).Scan(&displayName, &updatedAt); err != nil {
		t.Fatal(err)
	}
	if displayName != "New session name" || !updatedAt.Equal(now) {
		t.Errorf("new-session synchronization=(%q, %s), want updated values", displayName, updatedAt)
	}

	if err := repository.SynchronizeUser(t.Context(), SignedUser{ID: uuid.Nil}, now); err != nil {
		t.Fatalf("permissive synchronization rejected zero UUID and empty display name: %v", err)
	}

	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.LockUser(ctx, userID)
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Pool.Exec(t.Context(), `update users set deletion_locked_at = $1 where id = $2`, now, userID); err != nil {
		t.Fatal(err)
	}
	if err := repository.SynchronizeUser(t.Context(), user, now); !errors.Is(err, ErrAccountDeletionInProgress) {
		t.Errorf("synchronize locked user error=%v, want account deletion conflict", err)
	}
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.LockUser(ctx, userID)
	}); !errors.Is(err, ErrAccountDeletionInProgress) {
		t.Errorf("lock deleting user error=%v, want account deletion conflict", err)
	}

	missingID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	if err := postgres.RunInTransaction(t.Context(), db.Pool, func(ctx context.Context) error {
		return repository.LockUser(ctx, missingID)
	}); !errors.Is(err, ErrLocalUserNotFound) {
		t.Errorf("lock missing user error=%v, want local user not found", err)
	}
}
