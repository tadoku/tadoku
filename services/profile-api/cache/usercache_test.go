package cache_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
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

type mockSuppressionRepository struct {
	mu          sync.Mutex
	identityIDs []uuid.UUID
	err         error
	calls       int
}

func (m *mockSuppressionRepository) ListAccountDeletionSuppressedIdentityIDs(context.Context) ([]uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	if m.err != nil {
		return nil, m.err
	}
	return append([]uuid.UUID(nil), m.identityIDs...), nil
}

func (m *mockSuppressionRepository) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func containsUser(users []domain.UserCacheEntry, identityID uuid.UUID) bool {
	for _, user := range users {
		if user.ID == identityID.String() {
			return true
		}
	}
	return false
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

	c := cache.NewUserCache(kratos, &mockSuppressionRepository{}, time.Hour)
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

func TestUserCacheFiltersDurableSuppressionsOnStartupAndRestart(t *testing.T) {
	suppressedID := uuid.New()
	visibleID := uuid.New()
	kratos := &mockKratosClient{pages: map[int64]*domain.ListIdentitiesResult{
		0: {
			Identities: []domain.IdentityInfo{
				{ID: suppressedID.String(), DisplayName: "Deleted user"},
				{ID: visibleID.String(), DisplayName: "Visible user"},
			},
		},
	}}
	suppressions := &mockSuppressionRepository{identityIDs: []uuid.UUID{suppressedID}}

	for run := 0; run < 2; run++ {
		c := cache.NewUserCache(kratos, suppressions, time.Hour)
		c.Start()
		require.Eventually(t, func() bool {
			return len(c.GetUsers()) == 1
		}, time.Second, 10*time.Millisecond)
		users := c.GetUsers()
		assert.False(t, containsUser(users, suppressedID))
		assert.True(t, containsUser(users, visibleID))
		c.Stop()
	}
}

type blockingRefreshKratosClient struct {
	mu             sync.Mutex
	calls          int
	identities     []domain.IdentityInfo
	refreshStarted chan struct{}
	releaseRefresh chan struct{}
}

func (m *blockingRefreshKratosClient) ListIdentities(ctx context.Context, _ int64, _ int64) (*domain.ListIdentitiesResult, error) {
	m.mu.Lock()
	m.calls++
	call := m.calls
	m.mu.Unlock()
	if call == 2 {
		close(m.refreshStarted)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-m.releaseRefresh:
		}
	}
	return &domain.ListIdentitiesResult{Identities: m.identities}, nil
}

func TestUserCacheImmediateSuppressionWinsAgainstInflightRefresh(t *testing.T) {
	suppressedID := uuid.New()
	visibleID := uuid.New()
	kratos := &blockingRefreshKratosClient{
		identities: []domain.IdentityInfo{
			{ID: suppressedID.String(), DisplayName: "Deleted user"},
			{ID: visibleID.String(), DisplayName: "Visible user"},
		},
		refreshStarted: make(chan struct{}),
		releaseRefresh: make(chan struct{}),
	}
	suppressions := &mockSuppressionRepository{}
	c := cache.NewUserCache(kratos, suppressions, 10*time.Millisecond)
	c.Start()
	defer c.Stop()
	require.Eventually(t, func() bool { return len(c.GetUsers()) == 2 }, time.Second, 10*time.Millisecond)

	select {
	case <-kratos.refreshStarted:
	case <-time.After(time.Second):
		require.Fail(t, "refresh did not start")
	}
	c.SuppressAndEvict(suppressedID)
	assert.False(t, containsUser(c.GetUsers(), suppressedID))
	close(kratos.releaseRefresh)

	require.Eventually(t, func() bool { return suppressions.callCount() >= 2 }, time.Second, 10*time.Millisecond)
	assert.False(t, containsUser(c.GetUsers(), suppressedID))
	assert.True(t, containsUser(c.GetUsers(), visibleID))
}

func TestUserCacheFailsClosedWhenSuppressionLookupFails(t *testing.T) {
	identityID := uuid.New()
	kratos := &mockKratosClient{pages: map[int64]*domain.ListIdentitiesResult{
		0: {Identities: []domain.IdentityInfo{{ID: identityID.String(), DisplayName: "Visible user"}}},
	}}
	suppressions := &mockSuppressionRepository{}
	c := cache.NewUserCache(kratos, suppressions, 10*time.Millisecond)
	c.Start()
	defer c.Stop()
	require.Eventually(t, func() bool { return len(c.GetUsers()) == 1 }, time.Second, 10*time.Millisecond)

	suppressions.mu.Lock()
	suppressions.err = errors.New("database unavailable")
	suppressions.mu.Unlock()
	require.Eventually(t, func() bool {
		return suppressions.callCount() >= 2 && len(c.GetUsers()) == 0
	}, time.Second, 10*time.Millisecond)

	assert.Empty(t, c.GetUsers())
}
