package cache

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/profile-api/domain"
)

// AccountDeletionSuppressionRepository provides the durable set of accepted
// deletions that must never appear in the user cache.
type AccountDeletionSuppressionRepository interface {
	ListAccountDeletionSuppressedIdentityIDs(context.Context) ([]uuid.UUID, error)
}

type UserCache struct {
	mu                    sync.RWMutex
	users                 []domain.UserCacheEntry
	suppressedIdentityIDs map[string]struct{}
	kratos                domain.KratosClient
	suppressionRepository AccountDeletionSuppressionRepository
	refresh               time.Duration
	cancel                context.CancelFunc
}

func NewUserCache(
	kratos domain.KratosClient,
	suppressionRepository AccountDeletionSuppressionRepository,
	refresh time.Duration,
) *UserCache {
	return &UserCache{
		kratos:                kratos,
		suppressionRepository: suppressionRepository,
		refresh:               refresh,
		users:                 []domain.UserCacheEntry{},
		suppressedIdentityIDs: make(map[string]struct{}),
	}
}

func (c *UserCache) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go c.run(ctx)
}

func (c *UserCache) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *UserCache) run(ctx context.Context) {
	// Initial load with retry
	retries := 3
	for i := 0; i < retries; i++ {
		if err := c.refreshUsers(ctx); err != nil {
			log.Printf("UserCache: initial load attempt %d/%d failed: %v", i+1, retries, err)
			if i < retries-1 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(i+1) * 5 * time.Second):
				}
			}
		} else {
			break
		}
	}

	ticker := time.NewTicker(c.refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("UserCache: shutting down")
			return
		case <-ticker.C:
			if err := c.refreshUsers(ctx); err != nil {
				log.Printf("UserCache: refresh failed: %v", err)
			}
		}
	}
}

func (c *UserCache) refreshUsers(ctx context.Context) error {
	var allUsers []domain.UserCacheEntry
	seen := make(map[string]bool)
	page := int64(0)
	perPage := int64(500)

	for {
		result, err := c.kratos.ListIdentities(ctx, perPage, page)
		if err != nil {
			return err
		}

		for _, identity := range result.Identities {
			if seen[identity.ID] {
				continue
			}
			seen[identity.ID] = true
			allUsers = append(allUsers, domain.UserCacheEntry{
				ID:          identity.ID,
				DisplayName: identity.DisplayName,
				Email:       identity.Email,
				CreatedAt:   identity.CreatedAt,
			})
		}

		if !result.HasMore {
			break
		}
		page++
	}

	suppressedIdentityIDs, err := c.suppressionRepository.ListAccountDeletionSuppressedIdentityIDs(ctx)
	if err != nil {
		c.mu.Lock()
		c.users = []domain.UserCacheEntry{}
		c.mu.Unlock()
		return err
	}

	c.mu.Lock()
	for _, identityID := range suppressedIdentityIDs {
		c.suppressedIdentityIDs[identityID.String()] = struct{}{}
	}
	visibleUsers := make([]domain.UserCacheEntry, 0, len(allUsers))
	for _, user := range allUsers {
		if _, suppressed := c.suppressedIdentityIDs[user.ID]; suppressed {
			continue
		}
		visibleUsers = append(visibleUsers, user)
	}
	c.users = visibleUsers
	c.mu.Unlock()

	if len(visibleUsers) > 20000 {
		log.Printf("UserCache: WARNING - cache contains %d users, consider alternative approach", len(visibleUsers))
	}
	log.Printf("UserCache: refreshed with %d users", len(visibleUsers))
	return nil
}

// SuppressAndEvict keeps an accepted deletion out of the cache immediately and
// prevents an in-flight or later Kratos refresh from reintroducing it.
func (c *UserCache) SuppressAndEvict(identityID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := identityID.String()
	c.suppressedIdentityIDs[id] = struct{}{}
	visibleUsers := make([]domain.UserCacheEntry, 0, len(c.users))
	for _, user := range c.users {
		if user.ID != id {
			visibleUsers = append(visibleUsers, user)
		}
	}
	c.users = visibleUsers
}

// GetUsers returns a copy of all cached users
func (c *UserCache) GetUsers() []domain.UserCacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]domain.UserCacheEntry, len(c.users))
	copy(result, c.users)
	return result
}
