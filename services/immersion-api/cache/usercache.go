package cache

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sahilm/fuzzy"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

type UserCache struct {
	mu      sync.RWMutex
	users   []query.UserEntry
	kratos  query.KratosClient
	refresh time.Duration
}

func NewUserCache(kratos query.KratosClient, refresh time.Duration) *UserCache {
	return &UserCache{
		kratos:  kratos,
		refresh: refresh,
		users:   []query.UserEntry{},
	}
}

func (c *UserCache) Start(ctx context.Context) {
	// Initial load
	if err := c.refreshUsers(ctx); err != nil {
		log.Printf("UserCache: initial load failed: %v", err)
	}

	ticker := time.NewTicker(c.refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.refreshUsers(ctx); err != nil {
				log.Printf("UserCache: refresh failed: %v", err)
			}
		}
	}
}

func (c *UserCache) refreshUsers(ctx context.Context) error {
	var allUsers []query.UserEntry
	page := int64(0)
	perPage := int64(500)

	for {
		result, err := c.kratos.ListIdentities(ctx, perPage, page)
		if err != nil {
			return err
		}

		for _, identity := range result.Identities {
			allUsers = append(allUsers, query.UserEntry{
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

	log.Printf("UserCache: refreshed with %d users", len(allUsers))
	return nil
}

// userSearchSource implements fuzzy.Source for fuzzy matching on users
type userSearchSource struct {
	users []query.UserEntry
}

func (s userSearchSource) String(i int) string {
	u := s.users[i]
	return strings.ToLower(u.DisplayName + " " + u.Email)
}

func (s userSearchSource) Len() int {
	return len(s.users)
}

// Search performs fuzzy search on display name and email
// Returns matching users sorted by match score, with pagination and total count
func (c *UserCache) Search(queryStr string, limit, offset int) ([]query.UserEntry, int) {
	c.mu.RLock()
	users := c.users
	c.mu.RUnlock()

	if queryStr == "" {
		// No search query - return paginated results
		total := len(users)
		start := offset
		if start >= total {
			return []query.UserEntry{}, total
		}
		end := start + limit
		if end > total {
			end = total
		}
		return users[start:end], total
	}

	// Fuzzy search
	source := userSearchSource{users: users}
	matches := fuzzy.FindFrom(strings.ToLower(queryStr), source)

	// Limit to first page worth of results to avoid processing too many
	total := len(matches)
	if total > limit {
		matches = matches[:limit]
	}

	// Extract matched users in score order
	matchedUsers := make([]query.UserEntry, 0, len(matches))
	for _, match := range matches {
		matchedUsers = append(matchedUsers, users[match.Index])
	}

	return matchedUsers, total
}
