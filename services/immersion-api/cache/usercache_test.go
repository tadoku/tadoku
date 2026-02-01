package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/immersion-api/cache"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"

	"github.com/google/uuid"
)

// mockKratosClient implements query.KratosClient for testing
type mockKratosClient struct {
	identities []query.IdentityInfo
	err        error
	callCount  int
}

func (m *mockKratosClient) FetchIdentity(ctx context.Context, id uuid.UUID) (*query.UserTraits, error) {
	return nil, nil // Not used by cache
}

func (m *mockKratosClient) ListIdentities(ctx context.Context, perPage int64, page int64) (*query.ListIdentitiesResult, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}

	start := int(page * perPage)
	if start >= len(m.identities) {
		return &query.ListIdentitiesResult{Identities: nil, HasMore: false}, nil
	}

	end := start + int(perPage)
	hasMore := end < len(m.identities)
	if end > len(m.identities) {
		end = len(m.identities)
	}

	return &query.ListIdentitiesResult{
		Identities: m.identities[start:end],
		HasMore:    hasMore,
	}, nil
}

// testUser creates a test IdentityInfo with a display name and email derived from it
func testUser(displayName string) query.IdentityInfo {
	return query.IdentityInfo{
		ID:          uuid.New().String(),
		DisplayName: displayName,
		Email:       displayName + "@test.com",
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
}

// testUsers creates multiple test users from display names
func testUsers(names ...string) []query.IdentityInfo {
	users := make([]query.IdentityInfo, len(names))
	for i, name := range names {
		users[i] = testUser(name)
	}
	return users
}

// createCacheWithUsers creates a UserCache with the given identities, starts it, waits for load, and stops it
func createCacheWithUsers(t *testing.T, identities []query.IdentityInfo) *cache.UserCache {
	mock := &mockKratosClient{identities: identities}
	c := cache.NewUserCache(mock, time.Hour) // long refresh to prevent auto-refresh during test
	c.Start()
	// Give the cache time to load
	time.Sleep(50 * time.Millisecond)
	t.Cleanup(func() { c.Stop() })
	return c
}

func TestSearch_EmptyQuery_ReturnsPaginatedResults(t *testing.T) {
	users := testUsers("Alice", "Bob", "Charlie", "Diana", "Eve")
	c := createCacheWithUsers(t, users)

	// First page
	results, total := c.Search("", 2, 0)
	assert.Equal(t, 5, total)
	assert.Len(t, results, 2)
	assert.Equal(t, "Alice", results[0].DisplayName)
	assert.Equal(t, "Bob", results[1].DisplayName)

	// Second page
	results, total = c.Search("", 2, 2)
	assert.Equal(t, 5, total)
	assert.Len(t, results, 2)
	assert.Equal(t, "Charlie", results[0].DisplayName)
	assert.Equal(t, "Diana", results[1].DisplayName)

	// Third page (partial)
	results, total = c.Search("", 2, 4)
	assert.Equal(t, 5, total)
	assert.Len(t, results, 1)
	assert.Equal(t, "Eve", results[0].DisplayName)

	// Past end
	results, total = c.Search("", 2, 10)
	assert.Equal(t, 5, total)
	assert.Len(t, results, 0)
}

func TestSearch_FuzzyMatchDisplayName(t *testing.T) {
	users := testUsers("Alice", "Bob", "Zara", "Alice Smith")
	c := createCacheWithUsers(t, users)

	results, total := c.Search("alice", 10, 0)
	assert.GreaterOrEqual(t, total, 2)
	assert.GreaterOrEqual(t, len(results), 2)

	// Both Alice entries should be in the results (they match best)
	names := make([]string, len(results))
	for i, r := range results {
		names[i] = r.DisplayName
	}
	assert.Contains(t, names, "Alice")
	assert.Contains(t, names, "Alice Smith")
}

func TestSearch_FuzzyMatchEmail(t *testing.T) {
	users := testUsers("Alice", "Bob", "Zara")
	c := createCacheWithUsers(t, users)

	// Search for a unique email pattern
	results, total := c.Search("alice@test", 10, 0)
	assert.GreaterOrEqual(t, total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
	// Alice should be the best match
	assert.Equal(t, "Alice", results[0].DisplayName)
	assert.Equal(t, "Alice@test.com", results[0].Email)
}

func TestSearch_FuzzyWithPagination(t *testing.T) {
	// Create users where several match "test"
	users := testUsers("TestUser1", "TestUser2", "TestUser3", "TestUser4", "TestUser5", "Other", "Another")
	c := createCacheWithUsers(t, users)

	// First page of matches
	results, total := c.Search("testuser", 2, 0)
	assert.Equal(t, 5, total) // 5 users match "testuser"
	assert.Len(t, results, 2)

	// Second page
	results, total = c.Search("testuser", 2, 2)
	assert.Equal(t, 5, total)
	assert.Len(t, results, 2)

	// Third page (partial)
	results, total = c.Search("testuser", 2, 4)
	assert.Equal(t, 5, total)
	assert.Len(t, results, 1)
}

func TestSearch_NoMatches(t *testing.T) {
	users := testUsers("Alice", "Bob", "Charlie")
	c := createCacheWithUsers(t, users)

	results, total := c.Search("xyz123nonexistent", 10, 0)
	assert.Equal(t, 0, total)
	assert.Len(t, results, 0)
}

func TestSearch_EmptyCache(t *testing.T) {
	c := createCacheWithUsers(t, []query.IdentityInfo{})

	results, total := c.Search("", 10, 0)
	assert.Equal(t, 0, total)
	assert.Len(t, results, 0)

	results, total = c.Search("alice", 10, 0)
	assert.Equal(t, 0, total)
	assert.Len(t, results, 0)
}

func TestSearch_CaseInsensitive(t *testing.T) {
	users := testUsers("Alice", "ALICE", "alice")
	c := createCacheWithUsers(t, users)

	// Search with different cases should find all
	results, total := c.Search("ALICE", 10, 0)
	assert.Equal(t, 3, total)
	assert.Len(t, results, 3)

	results, total = c.Search("alice", 10, 0)
	assert.Equal(t, 3, total)
	assert.Len(t, results, 3)
}

func TestSearch_PartialMatch(t *testing.T) {
	users := testUsers("Alexander", "Alexandra", "Alex")
	c := createCacheWithUsers(t, users)

	results, total := c.Search("alex", 10, 0)
	assert.Equal(t, 3, total)
	assert.Len(t, results, 3)
}
