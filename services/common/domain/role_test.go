package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tadoku/tadoku/services/common/domain"
)

func TestIsRole_Match(t *testing.T) {
	ctx := context.WithValue(context.Background(), domain.CtxSessionKey, &domain.UserIdentity{Role: domain.RoleAdmin})
	assert.True(t, domain.IsRole(ctx, domain.RoleAdmin))
}

func TestIsRole_NoMatch(t *testing.T) {
	ctx := context.WithValue(context.Background(), domain.CtxSessionKey, &domain.UserIdentity{Role: domain.RoleUser})
	assert.False(t, domain.IsRole(ctx, domain.RoleAdmin))
}

func TestIsRole_NoSession(t *testing.T) {
	ctx := context.WithValue(context.Background(), domain.CtxSessionKey, nil)
	assert.False(t, domain.IsRole(ctx, domain.RoleUser))
}
