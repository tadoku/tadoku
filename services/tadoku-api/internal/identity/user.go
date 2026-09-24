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

func (u *User) UUID() (uuid.UUID, error) {
	if u == nil {
		return uuid.Nil, errx.NewInternalError("invalid signed user identity")
	}
	userID, err := uuid.Parse(u.Subject)
	if err != nil {
		return uuid.Nil, errx.NewInternalError("invalid signed user identity")
	}
	return userID, nil
}
