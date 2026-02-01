package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/cache"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

// mockKratosClient implements domain.KratosClient for testing
type mockKratosClient struct {
	identities []domain.IdentityInfo
	err        error
	callCount  int
}

func (m *mockKratosClient) FetchIdentity(ctx context.Context, id uuid.UUID) (*domain.UserTraits, error) {
	return nil, nil // Not used by cache
}

func (m *mockKratosClient) ListIdentities(ctx context.Context, perPage int64, page int64) (*domain.ListIdentitiesResult, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}

	start := int(page * perPage)
	if start >= len(m.identities) {
		return &domain.ListIdentitiesResult{Identities: nil, HasMore: false}, nil
	}

	end := start + int(perPage)
	hasMore := end < len(m.identities)
	if end > len(m.identities) {
		end = len(m.identities)
	}

	return &domain.ListIdentitiesResult{
		Identities: m.identities[start:end],
		HasMore:    hasMore,
	}, nil
}

// testUser creates a test IdentityInfo with a display name and email derived from it
func testUser(displayName string) domain.IdentityInfo {
	return domain.IdentityInfo{
		ID:          uuid.New().String(),
		DisplayName: displayName,
		Email:       displayName + "@test.com",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
}

// testUsers creates multiple test users from display names
func testUsers(names ...string) []domain.IdentityInfo {
	users := make([]domain.IdentityInfo, len(names))
	for i, name := range names {
		users[i] = testUser(name)
	}
	return users
}

// createCacheWithUsers creates a UserCache with the given identities, starts it, waits for load, and stops it
func createCacheWithUsers(t *testing.T, identities []domain.IdentityInfo) *cache.UserCache {
	mock := &mockKratosClient{identities: identities}
	c := cache.NewUserCache(mock, time.Hour) // long refresh to prevent auto-refresh during test
	c.Start()
	// Give the cache time to load
	time.Sleep(50 * time.Millisecond)
	t.Cleanup(func() { c.Stop() })
	return c
}

func TestGetUsers_ReturnsUsers(t *testing.T) {
	users := testUsers("Alice", "Bob", "Charlie")
	c := createCacheWithUsers(t, users)

	result := c.GetUsers()
	assert.Len(t, result, 3)
	assert.Equal(t, "Alice", result[0].DisplayName)
	assert.Equal(t, "Bob", result[1].DisplayName)
	assert.Equal(t, "Charlie", result[2].DisplayName)
}

func TestGetUsers_EmptyCache(t *testing.T) {
	c := createCacheWithUsers(t, []domain.IdentityInfo{})

	result := c.GetUsers()
	assert.Len(t, result, 0)
}

func TestGetUsers_ReturnsCopy(t *testing.T) {
	users := testUsers("Alice", "Bob")
	c := createCacheWithUsers(t, users)

	result1 := c.GetUsers()
	result2 := c.GetUsers()

	// Modify result1 and verify result2 is unaffected
	result1[0].DisplayName = "Modified"
	assert.Equal(t, "Alice", result2[0].DisplayName)
}
