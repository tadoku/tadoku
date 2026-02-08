package authzctx

import (
	"context"

	"github.com/tadoku/tadoku/services/common/authz/roles"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// WithUserIdentity attaches a user identity to ctx (or context.Background if ctx is nil).
// It does not attach any role claims.
func WithUserIdentity(ctx context.Context, u *commondomain.UserIdentity) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if u == nil {
		return ctx
	}
	return context.WithValue(ctx, commondomain.CtxIdentityKey, u)
}

func Guest() context.Context {
	return WithUserIdentity(context.Background(), &commondomain.UserIdentity{Subject: "guest"})
}

func UserSubject(subject string) context.Context {
	ctx := WithUserIdentity(context.Background(), &commondomain.UserIdentity{Subject: subject})
	return roles.WithClaims(ctx, roles.Claims{
		Subject:       subject,
		Authenticated: true,
	})
}

func AdminSubject(subject string) context.Context {
	ctx := WithUserIdentity(context.Background(), &commondomain.UserIdentity{Subject: subject})
	return roles.WithClaims(ctx, roles.Claims{
		Subject:       subject,
		Authenticated: true,
		Admin:         true,
	})
}

func UserIdentity(subject, displayName string) context.Context {
	ctx := WithUserIdentity(context.Background(), &commondomain.UserIdentity{
		Subject:     subject,
		DisplayName: displayName,
	})
	return roles.WithClaims(ctx, roles.Claims{
		Subject:       subject,
		Authenticated: true,
	})
}

func AdminIdentity(subject, displayName string) context.Context {
	ctx := WithUserIdentity(context.Background(), &commondomain.UserIdentity{
		Subject:     subject,
		DisplayName: displayName,
	})
	return roles.WithClaims(ctx, roles.Claims{
		Subject:       subject,
		Authenticated: true,
		Admin:         true,
	})
}
