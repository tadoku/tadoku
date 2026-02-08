package domain_test

import (
	"context"

	"github.com/tadoku/tadoku/services/common/authz/roles"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

const testSubjectID = "11111111-1111-1111-1111-111111111111"

func ctxWithToken(token *commondomain.UserIdentity) context.Context {
	if token == nil {
		return context.Background()
	}

	return context.WithValue(context.Background(), commondomain.CtxIdentityKey, token)
}

func ctxWithGuest() context.Context {
	return ctxWithToken(&commondomain.UserIdentity{Subject: "guest"})
}

func ctxWithUser() context.Context { return ctxWithUserSubject(testSubjectID) }

func ctxWithAdmin() context.Context { return ctxWithAdminSubject(testSubjectID) }

func ctxWithUserSubject(subject string) context.Context {
	ctx := ctxWithToken(&commondomain.UserIdentity{Subject: subject})
	return roles.WithClaims(ctx, roles.Claims{
		Subject:       subject,
		Authenticated: true,
	})
}

func ctxWithAdminSubject(subject string) context.Context {
	ctx := ctxWithToken(&commondomain.UserIdentity{Subject: subject})
	return roles.WithClaims(ctx, roles.Claims{
		Subject:       subject,
		Authenticated: true,
		Admin:         true,
	})
}

func ctxWithUserIdentity(subject, displayName string) context.Context {
	ctx := ctxWithToken(&commondomain.UserIdentity{
		Subject:     subject,
		DisplayName: displayName,
	})
	return roles.WithClaims(ctx, roles.Claims{
		Subject:       subject,
		Authenticated: true,
	})
}

func ctxWithAdminIdentity(subject, displayName string) context.Context {
	ctx := ctxWithToken(&commondomain.UserIdentity{
		Subject:     subject,
		DisplayName: displayName,
	})
	return roles.WithClaims(ctx, roles.Claims{
		Subject:       subject,
		Authenticated: true,
		Admin:         true,
	})
}
