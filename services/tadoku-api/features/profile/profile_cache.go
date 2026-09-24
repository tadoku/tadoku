package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	kratosapi "github.com/ory/kratos-client-go"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

const (
	userCacheRefreshInterval = 5 * time.Minute
	identityPageSize         = int64(500)
	identityCreatedAtLayout  = "2006-01-02T15:04:05Z"
)

type UserCache struct {
	mu         sync.Mutex
	users      []CachedUser
	loaded     bool
	checkedAt  time.Time
	identities *kratosclient.Client
}

func NewUserCache(identities *kratosclient.Client) *UserCache {
	return &UserCache{
		users:      []CachedUser{},
		identities: identities,
	}
}

func (c *UserCache) refresh(ctx context.Context, checkedAt time.Time) error {
	users, err := c.listUsers(ctx)
	if err != nil {
		if c.loaded && ctx.Err() == nil {
			c.checkedAt = checkedAt
		}
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	c.users = users
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

func decodeIdentityTraits(traits any) (identityTraits, error) {
	encoded, err := json.Marshal(traits)
	if err != nil {
		return identityTraits{}, err
	}
	var decoded identityTraits
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return identityTraits{}, err
	}
	return decoded, nil
}

func cachedUser(identity kratosapi.Identity) (CachedUser, bool) {
	if identity.GetSchemaId() != "user" {
		return CachedUser{}, false
	}

	traits, err := decodeIdentityTraits(identity.GetTraits())
	if err != nil {
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
