package domain

// UserCacheEntry represents a cached user from the identity provider
type UserCacheEntry struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}
