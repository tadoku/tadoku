package domain

import (
	"context"
)

type SessionToken struct {
	Subject     string
	DisplayName string
	Email       string
	Role        Role
}

func ParseSession(ctx context.Context) *SessionToken {
	if session, ok := ctx.Value(CtxSessionKey).(*SessionToken); ok {
		return session
	}

	return nil
}
