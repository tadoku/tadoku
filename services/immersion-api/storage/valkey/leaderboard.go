package valkey

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	valkeylib "github.com/valkey-io/valkey-go"
)

// Redis key prefixes for leaderboard sorted sets.
const (
	contestLeaderboardPrefix = "leaderboard:contest:"
	yearlyLeaderboardPrefix  = "leaderboard:yearly:"
	globalLeaderboardKey     = "leaderboard:global"
)

// incrementScript atomically checks if a sorted set exists and increments
// the member's score. Returns 1 if incremented, 0 if the key doesn't exist.
//
// KEYS[1] = sorted set key
// ARGV[1] = score increment (float)
// ARGV[2] = member (user ID)
var incrementScript = valkeylib.NewLuaScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then
  redis.call('ZINCRBY', KEYS[1], ARGV[1], ARGV[2])
  return 1
end
return 0
`)

// rebuildScript atomically deletes and repopulates a sorted set.
// Uses a single Lua script to ensure the operation is atomic — no partial
// state is visible to readers between delete and populate.
//
// KEYS[1] = sorted set key
// ARGV = alternating score, member pairs: [score1, member1, score2, member2, ...]
var rebuildScript = valkeylib.NewLuaScript(`
redis.call('DEL', KEYS[1])
if #ARGV > 0 then
  return redis.call('ZADD', KEYS[1], unpack(ARGV))
end
return 0
`)

// LeaderboardStore implements domain.LeaderboardStore using Valkey sorted sets.
type LeaderboardStore struct {
	client valkeylib.Client
}

// NewLeaderboardStore creates a new LeaderboardStore backed by the given Valkey client.
func NewLeaderboardStore(client valkeylib.Client) *LeaderboardStore {
	return &LeaderboardStore{client: client}
}

func contestLeaderboardKey(contestID uuid.UUID) string {
	return contestLeaderboardPrefix + contestID.String()
}

func yearlyLeaderboardKey(year int) string {
	return yearlyLeaderboardPrefix + strconv.Itoa(year)
}

func (s *LeaderboardStore) IncrementContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID, score float64) (bool, error) {
	return s.incrementScore(ctx, contestLeaderboardKey(contestID), userID.String(), score)
}

func (s *LeaderboardStore) IncrementYearlyScore(ctx context.Context, year int, userID uuid.UUID, score float64) (bool, error) {
	return s.incrementScore(ctx, yearlyLeaderboardKey(year), userID.String(), score)
}

func (s *LeaderboardStore) IncrementGlobalScore(ctx context.Context, userID uuid.UUID, score float64) (bool, error) {
	return s.incrementScore(ctx, globalLeaderboardKey, userID.String(), score)
}

func (s *LeaderboardStore) RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID, scores []domain.LeaderboardScore) error {
	return s.rebuildLeaderboard(ctx, contestLeaderboardKey(contestID), scores)
}

func (s *LeaderboardStore) RebuildYearlyLeaderboard(ctx context.Context, year int, scores []domain.LeaderboardScore) error {
	return s.rebuildLeaderboard(ctx, yearlyLeaderboardKey(year), scores)
}

func (s *LeaderboardStore) RebuildGlobalLeaderboard(ctx context.Context, scores []domain.LeaderboardScore) error {
	return s.rebuildLeaderboard(ctx, globalLeaderboardKey, scores)
}

// incrementScore atomically increments a user's score in a sorted set if the key exists.
func (s *LeaderboardStore) incrementScore(ctx context.Context, key string, member string, score float64) (bool, error) {
	scoreStr := strconv.FormatFloat(score, 'f', -1, 64)

	result, err := incrementScript.Exec(ctx, s.client, []string{key}, []string{scoreStr, member}).ToInt64()
	if err != nil {
		return false, fmt.Errorf("failed to increment leaderboard score for key %s: %w", key, err)
	}

	return result == 1, nil
}

// rebuildLeaderboard atomically replaces a sorted set with the given scores using a Lua script.
func (s *LeaderboardStore) rebuildLeaderboard(ctx context.Context, key string, scores []domain.LeaderboardScore) error {
	args := make([]string, 0, len(scores)*2)
	for _, entry := range scores {
		args = append(args, strconv.FormatFloat(entry.Score, 'f', -1, 64))
		args = append(args, entry.UserID.String())
	}

	err := rebuildScript.Exec(ctx, s.client, []string{key}, args).Error()
	if err != nil {
		return fmt.Errorf("failed to rebuild leaderboard for key %s: %w", key, err)
	}

	return nil
}
