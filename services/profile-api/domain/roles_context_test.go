package domain_test

import (
	"context"

	"github.com/tadoku/tadoku/services/common/testutil/authzctx"
)

const testSubjectID = "11111111-1111-1111-1111-111111111111"

func ctxWithUser() context.Context { return authzctx.UserSubject(testSubjectID) }

func ctxWithAdmin() context.Context { return authzctx.AdminSubject(testSubjectID) }
