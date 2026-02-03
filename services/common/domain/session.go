package domain

import "context"

// Deprecated: use UserIdentity instead. This alias exists for migration.
type SessionToken = UserIdentity

// ParseSession is deprecated, use ParseUserIdentity instead.
func ParseSession(ctx context.Context) *SessionToken {
	return ParseUserIdentity(ctx)
}
