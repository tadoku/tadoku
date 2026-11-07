package profile

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	valkeygo "github.com/valkey-io/valkey-go"
)

const avatarCacheTTL = 24 * time.Hour

type AvatarCache struct {
	client  valkeygo.Client
	timeout time.Duration
}

func NewAvatarCache(client valkeygo.Client, timeout time.Duration) *AvatarCache {
	return &AvatarCache{client: client, timeout: timeout}
}

// get reports cached false on a miss or a Valkey failure, and a nil avatar for a cached absence.
func (c *AvatarCache) get(ctx context.Context, userID uuid.UUID) (avatar *Avatar, cached bool) {
	key, ok := avatarCacheKey(ctx, userID)
	if !ok {
		return nil, false
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	value, err := c.client.Do(ctx, c.client.B().Get().Key(key).Build()).ToString()
	if err != nil {
		return nil, false
	}

	if value == "" {
		return nil, true
	}
	contentType, image, ok := strings.Cut(value, "\n")
	if !ok || !allowedAvatarTypes[contentType] {
		return nil, false
	}
	return &Avatar{ContentType: contentType, Image: []byte(image)}, true
}

// put stores a nil avatar as a cached absence. Failures are dropped: the next request fetches again.
func (c *AvatarCache) put(ctx context.Context, userID uuid.UUID, avatar *Avatar) {
	key, ok := avatarCacheKey(ctx, userID)
	if !ok {
		return
	}

	value := ""
	if avatar != nil {
		value = avatar.ContentType + "\n" + string(avatar.Image)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_ = c.client.Do(ctx, c.client.B().Set().Key(key).Value(value).Ex(avatarCacheTTL).Build()).Error()
}

func avatarCacheKey(ctx context.Context, userID uuid.UUID) (string, bool) {
	tenantKey, ok := tenant.FromContext(ctx)
	if !ok {
		return "", false
	}
	return "tenant:" + tenantKey.String() + ":profile:avatar:" + userID.String(), true
}
