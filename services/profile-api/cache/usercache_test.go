package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/profile-api/cache"
	"github.com/tadoku/tadoku/services/profile-api/domain"
)

type mockKratosClient struct {
	pages map[int64]*domain.ListIdentitiesResult
}

func (m *mockKratosClient) ListIdentities(ctx context.Context, perPage int64, page int64) (*domain.ListIdentitiesResult, error) {
	if result, ok := m.pages[page]; ok {
		return result, nil
	}
	return &domain.ListIdentitiesResult{Identities: nil, HasMore: false}, nil
}

func TestUserCache_DeduplicatesUsersAcrossPages(t *testing.T) {
	// Simulate pagination race condition: user "2" appears on both page 0 and page 1
	// This happens when a new user is created between page requests, shifting results
	kratos := &mockKratosClient{
		pages: map[int64]*domain.ListIdentitiesResult{
			0: {
				Identities: []domain.IdentityInfo{
					{ID: "1", DisplayName: "Alice", Email: "alice@test.com"},
					{ID: "2", DisplayName: "Bob", Email: "bob@test.com"},
				},
				HasMore: true,
			},
			1: {
				Identities: []domain.IdentityInfo{
					{ID: "2", DisplayName: "Bob", Email: "bob@test.com"}, // duplicate due to pagination shift
					{ID: "3", DisplayName: "Charlie", Email: "charlie@test.com"},
				},
				HasMore: false,
			},
		},
	}

	c := cache.NewUserCache(kratos, time.Hour)
	c.Start()
	defer c.Stop()

	// Wait for initial load
	require.Eventually(t, func() bool {
		return len(c.GetUsers()) > 0
	}, time.Second, 10*time.Millisecond, "cache should load users")

	users := c.GetUsers()

	// Should have 3 unique users, not 4
	assert.Len(t, users, 3)

	// Verify the correct users are present
	ids := make(map[string]bool)
	for _, u := range users {
		ids[u.ID] = true
	}
	assert.True(t, ids["1"], "should have user 1")
	assert.True(t, ids["2"], "should have user 2")
	assert.True(t, ids["3"], "should have user 3")
}
