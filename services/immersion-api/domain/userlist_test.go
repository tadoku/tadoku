package domain_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockUserListCache struct {
	users []domain.UserCacheEntry
}

func (m *mockUserListCache) GetUsers() []domain.UserCacheEntry {
	return m.users
}

func TestUserList_Execute(t *testing.T) {
	users := []domain.UserCacheEntry{
		{ID: "1", DisplayName: "Alice", Email: "alice@test.com"},
		{ID: "2", DisplayName: "Bob", Email: "bob@test.com"},
		{ID: "3", DisplayName: "Charlie", Email: "charlie@test.com"},
	}

	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache)

		result, err := svc.Execute(context.Background(), &domain.UserListRequest{})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("returns forbidden for non-admin", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache)

		ctx := context.WithValue(context.Background(), commondomain.CtxSessionKey, &commondomain.SessionToken{
			Role: commondomain.RoleUser,
		})

		result, err := svc.Execute(ctx, &domain.UserListRequest{})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, domain.ErrForbidden)
	})

	t.Run("returns all users for admin", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache)

		ctx := context.WithValue(context.Background(), commondomain.CtxSessionKey, &commondomain.SessionToken{
			Role: commondomain.RoleAdmin,
		})

		result, err := svc.Execute(ctx, &domain.UserListRequest{PerPage: 20})

		require.NoError(t, err)
		assert.Len(t, result.Users, 3)
		assert.Equal(t, 3, result.TotalSize)
	})

	t.Run("paginates results", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache)

		ctx := context.WithValue(context.Background(), commondomain.CtxSessionKey, &commondomain.SessionToken{
			Role: commondomain.RoleAdmin,
		})

		result, err := svc.Execute(ctx, &domain.UserListRequest{PerPage: 2, Page: 0})

		require.NoError(t, err)
		assert.Len(t, result.Users, 2)
		assert.Equal(t, 3, result.TotalSize)
		assert.Equal(t, "Alice", result.Users[0].DisplayName)
	})

	t.Run("filters by query with fuzzy search", func(t *testing.T) {
		cache := &mockUserListCache{users: users}
		svc := domain.NewUserList(cache)

		ctx := context.WithValue(context.Background(), commondomain.CtxSessionKey, &commondomain.SessionToken{
			Role: commondomain.RoleAdmin,
		})

		result, err := svc.Execute(ctx, &domain.UserListRequest{Query: "bob"})

		require.NoError(t, err)
		assert.Equal(t, 1, result.TotalSize)
		assert.Equal(t, "Bob", result.Users[0].DisplayName)
	})
}
