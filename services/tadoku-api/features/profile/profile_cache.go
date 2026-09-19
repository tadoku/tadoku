package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	kratosapi "github.com/ory/kratos-client-go"
)

const (
	userCacheRefreshInterval = 5 * time.Minute
	identityPageSize         = int64(500)
	identityCreatedAtLayout  = "2006-01-02T15:04:05Z"
)

type identityProvider interface {
	ListIdentities(context.Context, int64, string) ([]kratosapi.Identity, string, error)
}

type suppressionRepository interface {
	ListAccountDeletionSuppressedIdentityIDs(context.Context) ([]string, error)
}

type UserCache struct {
	mu                    sync.RWMutex
	users                 []CachedUser
	suppressedIdentityIDs map[string]struct{}
	identities            identityProvider
	suppressions          suppressionRepository
}

func NewUserCache(identities identityProvider, suppressions suppressionRepository) *UserCache {
	return &UserCache{
		users:                 []CachedUser{},
		suppressedIdentityIDs: make(map[string]struct{}),
		identities:            identities,
		suppressions:          suppressions,
	}
}

// Run refreshes the cache until ctx is canceled. The initial retries preserve
// the legacy cache's startup behavior; later failures wait for the next tick.
func (c *UserCache) Run(ctx context.Context) {
	for attempt := 0; attempt < 3; attempt++ {
		if err := c.Refresh(ctx); err == nil {
			break
		} else {
			slog.ErrorContext(ctx, "initial user cache refresh failed", "attempt", attempt+1, "error", err)
		}
		if attempt < 2 && !waitForUserCache(ctx, time.Duration(attempt+1)*5*time.Second) {
			return
		}
	}

	ticker := time.NewTicker(userCacheRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.Refresh(ctx); err != nil {
				slog.ErrorContext(ctx, "user cache refresh failed", "error", err)
			}
		}
	}
}

func waitForUserCache(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// Refresh synchronously replaces the visible snapshot after both providers
// succeed. Identity-provider failures retain the previous complete snapshot;
// suppression failures clear it so accepted deletions cannot reappear.
func (c *UserCache) Refresh(ctx context.Context) error {
	users, err := c.listUsers(ctx)
	if err != nil {
		return err
	}

	suppressedIDs, err := c.suppressions.ListAccountDeletionSuppressedIdentityIDs(ctx)
	if err != nil {
		c.mu.Lock()
		c.users = []CachedUser{}
		c.mu.Unlock()
		return fmt.Errorf("refresh user suppressions: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, identityID := range suppressedIDs {
		c.suppressedIdentityIDs[identityID] = struct{}{}
	}

	visible := make([]CachedUser, 0, len(users))
	for _, user := range users {
		if _, suppressed := c.suppressedIdentityIDs[user.ID]; !suppressed {
			visible = append(visible, user)
		}
	}
	c.users = visible
	return nil
}

func (c *UserCache) listUsers(ctx context.Context) ([]CachedUser, error) {
	users := make([]CachedUser, 0)
	seenIdentityIDs := make(map[string]struct{})
	seenPageTokens := make(map[string]struct{})
	pageToken := ""

	for {
		identities, nextPageToken, err := c.identities.ListIdentities(ctx, identityPageSize, pageToken)
		if err != nil {
			return nil, err
		}

		for _, identity := range identities {
			user, ok := cachedUser(identity)
			if !ok {
				continue
			}
			if _, duplicate := seenIdentityIDs[user.ID]; duplicate {
				continue
			}
			seenIdentityIDs[user.ID] = struct{}{}
			users = append(users, user)
		}

		if nextPageToken == "" {
			return users, nil
		}
		if _, repeated := seenPageTokens[nextPageToken]; repeated {
			return nil, fmt.Errorf("refresh users: repeated next page token")
		}
		seenPageTokens[nextPageToken] = struct{}{}
		pageToken = nextPageToken
	}
}

type identityTraits struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

func cachedUser(identity kratosapi.Identity) (CachedUser, bool) {
	if identity.GetSchemaId() != "user" {
		return CachedUser{}, false
	}

	encodedTraits, err := json.Marshal(identity.GetTraits())
	if err != nil {
		return CachedUser{}, false
	}
	var traits identityTraits
	if err := json.Unmarshal(encodedTraits, &traits); err != nil {
		return CachedUser{}, false
	}

	createdAt := ""
	if identity.CreatedAt != nil {
		createdAt = identity.GetCreatedAt().Format(identityCreatedAtLayout)
	}
	return CachedUser{
		ID:          identity.GetId(),
		DisplayName: traits.DisplayName,
		Email:       traits.Email,
		CreatedAt:   createdAt,
	}, true
}

func (c *UserCache) Users() []CachedUser {
	c.mu.RLock()
	defer c.mu.RUnlock()

	users := make([]CachedUser, len(c.users))
	copy(users, c.users)
	return users
}
