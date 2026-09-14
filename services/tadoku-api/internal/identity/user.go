// Package identity carries verified user claims across application layers.
package identity

import (
	"context"
	"time"
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
