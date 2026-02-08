package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockUserListCache struct {
	users []domain.UserCacheEntry
}

func (m *mockUserListCache) GetUsers() []domain.UserCacheEntry {
	return m.users
}

type mockRolesService struct {
	claimsBySubject map[string]commonroles.Claims
	err             error
}

func (m *mockRolesService) ClaimsForSubject(ctx context.Context, subjectID string) (commonroles.Claims, error) {
	if m.err != nil {
		return commonroles.Claims{Subject: subjectID, Authenticated: true, Err: m.err}, m.err
	}
	if m.claimsBySubject == nil {
		return commonroles.Claims{Subject: subjectID, Authenticated: true}, nil
	}
	if c, ok := m.claimsBySubject[subjectID]; ok {
		return c, nil
	}
	return commonroles.Claims{Subject: subjectID, Authenticated: true}, nil
}

func (m *mockRolesService) ClaimsForSubjects(ctx context.Context, subjectIDs []string) (map[string]commonroles.Claims, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make(map[string]commonroles.Claims, len(subjectIDs))
	for _, id := range subjectIDs {
		c, err := m.ClaimsForSubject(ctx, id)
		if err != nil {
			return nil, err
		}
		out[id] = c
	}
	return out, nil
}

func TestUserList_Execute(t *testing.T) {
	users := []domain.UserCacheEntry{
		{ID: "1", DisplayName: "Alice", Email: "alice@test.com"},
		{ID: "2", DisplayName: "Bob", Email: "bob@test.com"},
		{ID: "3", DisplayName: "Charlie", Email: "charlie@test.com"},
	}

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache, &mockRolesService{})

		result, err := svc.Execute(context.Background(), &domain.UserListRequest{})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("returns forbidden for non-admin", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache, &mockRolesService{})

		ctx := ctxWithUser()

		result, err := svc.Execute(ctx, &domain.UserListRequest{})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("returns all users for admin", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache, &mockRolesService{})

		ctx := ctxWithAdmin()

		result, err := svc.Execute(ctx, &domain.UserListRequest{PerPage: 20})

		require.NoError(t, err)
		assert.Len(t, result.Users, 3)
		assert.Equal(t, 3, result.TotalSize)
	})

	t.Run("paginates results", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache, &mockRolesService{})

		ctx := ctxWithAdmin()

		result, err := svc.Execute(ctx, &domain.UserListRequest{PerPage: 2, Page: 0})

		require.NoError(t, err)
		assert.Len(t, result.Users, 2)
		assert.Equal(t, 3, result.TotalSize)
		assert.Equal(t, "Alice", result.Users[0].DisplayName)
	})

	t.Run("filters by query with fuzzy search", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache, &mockRolesService{})

		ctx := ctxWithAdmin()

		result, err := svc.Execute(ctx, &domain.UserListRequest{Query: "bob"})

		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalSize)
		assert.Equal(t, "Bob", result.Users[0].DisplayName)
	})
}
