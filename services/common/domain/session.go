package domain

import (
	"context"
	"time"
)

type SessionToken struct {
	Subject     string
	DisplayName string
	Email       string
	Role        Role
	CreatedAt   time.Time
}

func ParseSession(ctx context.Context) *SessionToken {
	if session, ok := ctx.Value(CtxSessionKey).(*SessionToken); ok {
		return session
	}

	return nil
}
