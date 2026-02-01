package cache

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

type UserCache struct {
	mu      sync.RWMutex
	users   []domain.UserCacheEntry
	kratos  query.KratosClient
	refresh time.Duration
	cancel  context.CancelFunc
}

func NewUserCache(kratos query.KratosClient, refresh time.Duration) *UserCache {
	return &UserCache{
		kratos:  kratos,
		refresh: refresh,
		users:   []domain.UserCacheEntry{},
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
	page := int64(0)
	perPage := int64(500)

	for {
		result, err := c.kratos.ListIdentities(ctx, perPage, page)
		if err != nil {
			return err
		}

		for _, identity := range result.Identities {
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

	c.mu.Lock()
	c.users = allUsers
	c.mu.Unlock()

	if len(allUsers) > 20000 {
		log.Printf("UserCache: WARNING - cache contains %d users, consider alternative approach", len(allUsers))
	}
	log.Printf("UserCache: refreshed with %d users", len(allUsers))
	return nil
}

// GetUsers returns a copy of all cached users
func (c *UserCache) GetUsers() []domain.UserCacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]domain.UserCacheEntry, len(c.users))
	copy(result, c.users)
	return result
}
