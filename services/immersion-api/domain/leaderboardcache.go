package domain

import "context"

// LeaderboardCacheKind identifies one of the bounded leaderboard cache
// families. It is safe to use as a metrics label.
type LeaderboardCacheKind string

const (
	LeaderboardCacheKindGlobal  LeaderboardCacheKind = "global"
	LeaderboardCacheKindYearly  LeaderboardCacheKind = "yearly"
	LeaderboardCacheKindContest LeaderboardCacheKind = "contest"
)

// LeaderboardCacheOperation identifies the bounded cache operation being
// observed. It is safe to use as a metrics label.
type LeaderboardCacheOperation string

const (
	LeaderboardCacheOperationFetch   LeaderboardCacheOperation = "fetch"
	LeaderboardCacheOperationRebuild LeaderboardCacheOperation = "rebuild"
	LeaderboardCacheOperationUpdate  LeaderboardCacheOperation = "update"
)

// LeaderboardCacheOutcome identifies the bounded result of a cache operation.
// Fallback means the authoritative repository was used because the cache could
// not serve the request.
type LeaderboardCacheOutcome string

const (
	LeaderboardCacheOutcomeSuccess  LeaderboardCacheOutcome = "success"
	LeaderboardCacheOutcomeMiss     LeaderboardCacheOutcome = "miss"
	LeaderboardCacheOutcomeFailure  LeaderboardCacheOutcome = "failure"
	LeaderboardCacheOutcomeFallback LeaderboardCacheOutcome = "fallback"
)

// LeaderboardCacheObservation is the domain-to-observability contract for
// cache health. Err is diagnostic context for logs only and must never be used
// as a metrics label.
type LeaderboardCacheObservation struct {
	Kind      LeaderboardCacheKind
	Operation LeaderboardCacheOperation
	Outcome   LeaderboardCacheOutcome
	Err       error
}

// LeaderboardCacheObserver is defined where cache operations are consumed so
// the domain does not depend on a concrete observability implementation.
// It deliberately has no error return: telemetry must not affect requests.
type LeaderboardCacheObserver interface {
	ObserveLeaderboardCache(context.Context, LeaderboardCacheObservation)
}
