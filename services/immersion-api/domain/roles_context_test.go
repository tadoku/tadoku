package domain_test

import (
	"context"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/testutil/authzctx"
)

const testSubjectID = "11111111-1111-1111-1111-111111111111"

func ctxWithToken(token *commondomain.UserIdentity) context.Context {
	return authzctx.WithUserIdentity(context.Background(), token)
}

func ctxWithGuest() context.Context {
	return authzctx.Guest()
}

func ctxWithUser() context.Context { return ctxWithUserSubject(testSubjectID) }

func ctxWithAdmin() context.Context { return ctxWithAdminSubject(testSubjectID) }

func ctxWithUserSubject(subject string) context.Context {
	return authzctx.UserSubject(subject)
}

func ctxWithAdminSubject(subject string) context.Context {
	return authzctx.AdminSubject(subject)
}

func ctxWithUserIdentity(subject, displayName string) context.Context {
	return authzctx.UserIdentity(subject, displayName)
}

func ctxWithAdminIdentity(subject, displayName string) context.Context {
	return authzctx.AdminIdentity(subject, displayName)
}
