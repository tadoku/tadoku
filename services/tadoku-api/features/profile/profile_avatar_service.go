package profile

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
)

type Avatars struct {
	identities *identity.ReadService
	cache      *AvatarCache
	gravatar   *Gravatar
}

func NewAvatars(identities *identity.ReadService, cache *AvatarCache, gravatar *Gravatar) *Avatars {
	return &Avatars{identities: identities, cache: cache, gravatar: gravatar}
}

// Find returns ErrAvatarNotFound when Gravatar has no image for the user. It checks the user before the cache, so a
// deleted user's cached avatar is never served.
func (a *Avatars) Find(ctx context.Context, userID uuid.UUID) (Avatar, error) {
	account, err := a.identities.FindAccount(ctx, userID)
	if err != nil {
		return Avatar{}, err
	}

	avatar, cached := a.cache.get(ctx, userID)
	if !cached {
		avatar, err = a.gravatar.fetch(ctx, account.Email)
		if err != nil {
			return Avatar{}, err
		}
		a.cache.put(ctx, userID, avatar)
	}

	if avatar == nil {
		return Avatar{}, ErrAvatarNotFound
	}
	return *avatar, nil
}
