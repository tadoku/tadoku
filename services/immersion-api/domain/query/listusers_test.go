package query_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

// mockUserCache implements query.UserCache for testing
type mockUserCache struct {
	users []query.UserEntry
}

func (m *mockUserCache) GetUsers() []query.UserEntry {
	return m.users
}

// testUsers creates test UserEntry slices from display names
func testUsers(names ...string) []query.UserEntry {
	users := make([]query.UserEntry, len(names))
	for i, name := range names {
		users[i] = query.UserEntry{
			ID:          "id-" + name,
			DisplayName: name,
			Email:       name + "@test.com",
			CreatedAt:   "2024-01-01T00:00:00Z",
		}
	}
	return users
}

// adminContext returns a context with an admin session
func adminContext() context.Context {
	return context.WithValue(context.Background(), domain.CtxSessionKey, &domain.SessionToken{
		Role: domain.RoleAdmin,
	})
}

func TestListUsers_EmptyQuery_ReturnsPaginatedResults(t *testing.T) {
	cache := &mockUserCache{users: testUsers("Alice", "Bob", "Charlie", "Diana", "Eve")}
	svc := query.NewService(nil, nil, nil, cache)

	// First page
	resp, err := svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 2, Page: 0, Query: ""})
	assert.NoError(t, err)
	assert.Equal(t, 5, resp.TotalSize)
	assert.Len(t, resp.Users, 2)
	assert.Equal(t, "Alice", resp.Users[0].DisplayName)
	assert.Equal(t, "Bob", resp.Users[1].DisplayName)

	// Second page
	resp, err = svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 2, Page: 1, Query: ""})
	assert.NoError(t, err)
	assert.Equal(t, 5, resp.TotalSize)
	assert.Len(t, resp.Users, 2)
	assert.Equal(t, "Charlie", resp.Users[0].DisplayName)
	assert.Equal(t, "Diana", resp.Users[1].DisplayName)

	// Third page (partial)
	resp, err = svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 2, Page: 2, Query: ""})
	assert.NoError(t, err)
	assert.Equal(t, 5, resp.TotalSize)
	assert.Len(t, resp.Users, 1)
	assert.Equal(t, "Eve", resp.Users[0].DisplayName)

	// Past end
	resp, err = svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 2, Page: 10, Query: ""})
	assert.NoError(t, err)
	assert.Equal(t, 5, resp.TotalSize)
	assert.Len(t, resp.Users, 0)
}

func TestListUsers_FuzzyMatchDisplayName(t *testing.T) {
	cache := &mockUserCache{users: testUsers("Alice", "Bob", "Zara", "Alice Smith")}
	svc := query.NewService(nil, nil, nil, cache)

	resp, err := svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: "alice"})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, resp.TotalSize, 2)
	assert.GreaterOrEqual(t, len(resp.Users), 2)

	// Both Alice entries should be in the results
	names := make([]string, len(resp.Users))
	for i, u := range resp.Users {
		names[i] = u.DisplayName
	}
	assert.Contains(t, names, "Alice")
	assert.Contains(t, names, "Alice Smith")
}

func TestListUsers_FuzzyMatchEmail(t *testing.T) {
	cache := &mockUserCache{users: testUsers("Alice", "Bob", "Zara")}
	svc := query.NewService(nil, nil, nil, cache)

	resp, err := svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: "alice@test"})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, resp.TotalSize, 1)
	assert.GreaterOrEqual(t, len(resp.Users), 1)
	// Alice should be the best match
	assert.Equal(t, "Alice", resp.Users[0].DisplayName)
	assert.Equal(t, "Alice@test.com", resp.Users[0].Email)
}

func TestListUsers_FuzzyWithPagination(t *testing.T) {
	cache := &mockUserCache{users: testUsers("TestUser1", "TestUser2", "TestUser3", "TestUser4", "TestUser5", "Other", "Another")}
	svc := query.NewService(nil, nil, nil, cache)

	// First page of matches
	resp, err := svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 2, Page: 0, Query: "testuser"})
	assert.NoError(t, err)
	assert.Equal(t, 5, resp.TotalSize)
	assert.Len(t, resp.Users, 2)

	// Second page
	resp, err = svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 2, Page: 1, Query: "testuser"})
	assert.NoError(t, err)
	assert.Equal(t, 5, resp.TotalSize)
	assert.Len(t, resp.Users, 2)

	// Third page (partial)
	resp, err = svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 2, Page: 2, Query: "testuser"})
	assert.NoError(t, err)
	assert.Equal(t, 5, resp.TotalSize)
	assert.Len(t, resp.Users, 1)
}

func TestListUsers_NoMatches(t *testing.T) {
	cache := &mockUserCache{users: testUsers("Alice", "Bob", "Charlie")}
	svc := query.NewService(nil, nil, nil, cache)

	resp, err := svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: "xyz123nonexistent"})
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.TotalSize)
	assert.Len(t, resp.Users, 0)
}

func TestListUsers_EmptyCache(t *testing.T) {
	cache := &mockUserCache{users: []query.UserEntry{}}
	svc := query.NewService(nil, nil, nil, cache)

	resp, err := svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: ""})
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.TotalSize)
	assert.Len(t, resp.Users, 0)

	resp, err = svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: "alice"})
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.TotalSize)
	assert.Len(t, resp.Users, 0)
}

func TestListUsers_CaseInsensitive(t *testing.T) {
	cache := &mockUserCache{users: testUsers("Alice", "ALICE", "alice")}
	svc := query.NewService(nil, nil, nil, cache)

	// Search with different cases should find all
	resp, err := svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: "ALICE"})
	assert.NoError(t, err)
	assert.Equal(t, 3, resp.TotalSize)
	assert.Len(t, resp.Users, 3)

	resp, err = svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: "alice"})
	assert.NoError(t, err)
	assert.Equal(t, 3, resp.TotalSize)
	assert.Len(t, resp.Users, 3)
}

func TestListUsers_PartialMatch(t *testing.T) {
	cache := &mockUserCache{users: testUsers("Alexander", "Alexandra", "Alex")}
	svc := query.NewService(nil, nil, nil, cache)

	resp, err := svc.ListUsers(adminContext(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: "alex"})
	assert.NoError(t, err)
	assert.Equal(t, 3, resp.TotalSize)
	assert.Len(t, resp.Users, 3)
}

func TestListUsers_Unauthorized(t *testing.T) {
	cache := &mockUserCache{users: testUsers("Alice")}
	svc := query.NewService(nil, nil, nil, cache)

	// No session
	_, err := svc.ListUsers(context.Background(), &query.ListUsersRequest{PerPage: 10, Page: 0, Query: ""})
	assert.Equal(t, query.ErrUnauthorized, err)
}

func TestListUsers_Forbidden(t *testing.T) {
	cache := &mockUserCache{users: testUsers("Alice")}
	svc := query.NewService(nil, nil, nil, cache)

	// Non-admin user
	ctx := context.WithValue(context.Background(), domain.CtxSessionKey, &domain.SessionToken{
		Role: domain.RoleUser,
	})
	_, err := svc.ListUsers(ctx, &query.ListUsersRequest{PerPage: 10, Page: 0, Query: ""})
	assert.Equal(t, query.ErrForbidden, err)
}
