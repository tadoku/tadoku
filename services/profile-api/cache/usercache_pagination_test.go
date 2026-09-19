package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/profile-api/domain"
)

type paginatedKratosClient func(context.Context, int64, string) ([]domain.IdentityInfo, string, error)

func (f paginatedKratosClient) ListIdentities(ctx context.Context, size int64, token string) ([]domain.IdentityInfo, string, error) {
	return f(ctx, size, token)
}

type emptySuppressions struct{}

func (emptySuppressions) ListAccountDeletionSuppressedIdentityIDs(context.Context) ([]uuid.UUID, error) {
	return nil, nil
}

func TestUserCacheCursorPagination(t *testing.T) {
	const identityCount = 1205
	unavailable := errors.New("Kratos unavailable")
	for _, scenario := range []string{"all pages", "later page fails", "repeated cursor"} {
		t.Run(scenario, func(t *testing.T) {
			requests := 0
			client := paginatedKratosClient(func(ctx context.Context, size int64, token string) ([]domain.IdentityInfo, string, error) {
				requests++
				if size != 500 {
					t.Errorf("page size = %d, want cache's chosen size 500", size)
				}
				if requests > 3 {
					return nil, "", errors.New("pagination did not stop")
				}
				var start, end int
				var next string
				switch token {
				case "":
					start, end, next = 0, 500, "after-500"
				case "after-500":
					if scenario == "later page fails" {
						return nil, "", unavailable
					}
					start, end, next = 500, 1000, "after-1000"
					if scenario == "repeated cursor" {
						next = token
					}
				case "after-1000":
					start, end = 1000, identityCount
				default:
					return nil, "", fmt.Errorf("unexpected token %q", token)
				}
				identities := make([]domain.IdentityInfo, 0, end-start)
				for i := start; i < end; i++ {
					identities = append(identities, domain.IdentityInfo{ID: fmt.Sprintf("identity-%d", i)})
				}
				return identities, next, nil
			})
			c := NewUserCache(client, emptySuppressions{}, time.Hour)
			c.users = []domain.UserCacheEntry{{ID: "previous-refresh"}}
			err := c.refreshUsers(context.Background())
			users := c.GetUsers()
			if scenario != "all pages" {
				if err == nil {
					t.Fatal("refresh succeeded, want pagination failure")
				}
				if scenario == "later page fails" && !errors.Is(err, unavailable) {
					t.Errorf("error = %v, want dependency failure", err)
				}
				if requests != 2 {
					t.Errorf("requests = %d, want 2", requests)
				}
				if len(users) != 1 || users[0].ID != "previous-refresh" {
					t.Fatal("failed refresh replaced cache with partial results")
				}
				return
			}
			if err != nil {
				t.Fatalf("refresh: %v", err)
			}
			if len(users) != identityCount || requests != 3 {
				t.Fatalf("users=%d requests=%d, want %d and 3", len(users), requests, identityCount)
			}
			for i, user := range users {
				if user.ID != fmt.Sprintf("identity-%d", i) {
					t.Fatalf("user %d = %q, want matching identity", i, user.ID)
				}
			}
		})
	}
}
