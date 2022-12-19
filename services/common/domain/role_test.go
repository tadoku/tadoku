package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tadoku/tadoku/services/common/domain"
)

type gettable struct {
	m map[string]interface{}
}

func (g *gettable) Get(key string) interface{} {
	return g.m[key]
}

func NewGettable(key string, val interface{}) domain.Gettable {
	return &gettable{m: map[string]interface{}{key: val}}
}

func TestIsRole_Match(t *testing.T) {
	ctx := NewGettable("session", &domain.SessionToken{Role: domain.RoleAdmin})
	assert.True(t, domain.IsRole(ctx, domain.RoleAdmin))
}

func TestIsRole_NoMatch(t *testing.T) {
	ctx := NewGettable("session", &domain.SessionToken{Role: domain.RoleUser})
	assert.False(t, domain.IsRole(ctx, domain.RoleAdmin))
}

func TestIsRole_NoSession(t *testing.T) {
	ctx := NewGettable("session", nil)
	assert.False(t, domain.IsRole(ctx, domain.RoleUser))
}
