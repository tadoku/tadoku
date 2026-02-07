package domain_test

import (
	"context"

	"github.com/tadoku/tadoku/services/common/authz/roles"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

const testSubjectID = "11111111-1111-1111-1111-111111111111"

func ctxWithRole(role commondomain.Role) context.Context {
	return ctxWithToken(&commondomain.UserIdentity{
		Subject: testSubjectID,
		Role:    role,
	})
}

func ctxWithToken(token *commondomain.UserIdentity) context.Context {
	if token == nil {
		return context.Background()
	}

	ctx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, token)

	// Guest tokens are treated as unauthenticated and do not get role claims.
	if token.Role == commondomain.RoleGuest || token.Subject == "guest" {
		return ctx
	}

	claims := roles.Claims{
		Subject:       token.Subject,
		Authenticated: true,
		Admin:         token.Role == commondomain.RoleAdmin,
		Banned:        token.Role == commondomain.RoleBanned,
	}

	return roles.WithClaims(ctx, claims)
}

