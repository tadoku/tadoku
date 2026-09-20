package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	kratosapi "github.com/ory/kratos-client-go"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
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
	mu        sync.Mutex
	users     []CachedUser
	loaded    bool
	checkedAt time.Time
	// Accepted account deletions must disappear from administrator listings
	// before Kratos deletion completes. Learned IDs stay suppressed for this
	// cache's lifetime so a later provider snapshot cannot reintroduce them.
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

// refresh is called with c.mu held. It replaces the visible snapshot after
// both providers succeed. Identity-provider failures retain the previous
// complete snapshot; suppression failures clear it so accepted deletions
// cannot reappear.
func (c *UserCache) refresh(ctx context.Context, checkedAt time.Time) error {
	users, err := c.listUsers(ctx)
	if err != nil {
		if c.loaded && ctx.Err() == nil {
			c.checkedAt = checkedAt
		}
		return err
	}

	suppressedIDs, err := c.suppressions.ListAccountDeletionSuppressedIdentityIDs(ctx)
	if err != nil {
		c.users = []CachedUser{}
		c.loaded = true
		if ctx.Err() == nil {
			c.checkedAt = checkedAt
		}
		return fmt.Errorf("refresh user suppressions: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

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
	c.loaded = true
	c.checkedAt = checkedAt
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

func (c *UserCache) Users(ctx context.Context) ([]CachedUser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	now := timex.Now()
	if c.loaded && now.Before(c.checkedAt.Add(userCacheRefreshInterval)) {
		return append([]CachedUser{}, c.users...), nil
	}

	err := c.refresh(ctx, now)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		if !c.loaded {
			return nil, err
		}
		slog.ErrorContext(ctx, "user cache refresh failed; serving current snapshot", "error", err)
	}
	return append([]CachedUser{}, c.users...), nil
}
