package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

// User includes the signed guest subject used for anonymous gateway requests.
// CreatedAt is the token's issued-at time, not an account creation timestamp.
type User struct {
	Subject     string
	DisplayName string
	Email       string
	CreatedAt   time.Time
}

type userKey struct{}

func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func FromContext(ctx context.Context) *User {
	user, _ := ctx.Value(userKey{}).(*User)
	return user
}

func ActorID(ctx context.Context) (uuid.UUID, bool) {
	user := FromContext(ctx)
	if user == nil {
		return uuid.Nil, false
	}
	userID, err := uuid.Parse(user.Subject)
	if err != nil || userID == uuid.Nil {
		return uuid.Nil, false
	}
	return userID, true
}

func RequireActorID(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ActorID(ctx)
	if !ok {
		return uuid.Nil, errx.NewUnauthorizedError("unauthorized")
	}
	return userID, nil
}
