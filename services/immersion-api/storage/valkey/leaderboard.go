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

// updateScoreScript atomically checks if a sorted set exists and sets
// the member's absolute score via ZADD. Returns 1 if updated, 0 if
// the key doesn't exist.
//
// KEYS[1] = sorted set key
// ARGV[1] = score (float)
// ARGV[2] = member (user ID)
var updateScoreScript = valkeylib.NewLuaScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then
  redis.call('ZADD', KEYS[1], ARGV[1], ARGV[2])
  return 1
end
return 0
`)

// updateOfficialScoresScript atomically checks if yearly and global sorted
// sets exist and sets the member's absolute scores via ZADD. Returns a
// bitmask: bit 0 = yearly updated, bit 1 = global updated.
//
// KEYS[1] = yearly sorted set key
// KEYS[2] = global sorted set key
// ARGV[1] = yearly score (float)
// ARGV[2] = global score (float)
// ARGV[3] = member (user ID)
var updateOfficialScoresScript = valkeylib.NewLuaScript(`
local result = 0
if redis.call('EXISTS', KEYS[1]) == 1 then
  redis.call('ZADD', KEYS[1], ARGV[1], ARGV[3])
  result = result + 1
end
if redis.call('EXISTS', KEYS[2]) == 1 then
  redis.call('ZADD', KEYS[2], ARGV[2], ARGV[3])
  result = result + 2
end
return result
`)

// rebuildScript atomically deletes and repopulates a sorted set.
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

// rebuildOfficialScript atomically deletes and repopulates both yearly and
// global sorted sets.
//
// KEYS[1] = yearly sorted set key
// KEYS[2] = global sorted set key
// ARGV[1] = number of yearly score/member pairs (N)
// ARGV[2..2*N+1] = yearly score, member pairs
// ARGV[2*N+2..end] = global score, member pairs
var rebuildOfficialScript = valkeylib.NewLuaScript(`
redis.call('DEL', KEYS[1])
redis.call('DEL', KEYS[2])
local yearlyCount = tonumber(ARGV[1])
if yearlyCount > 0 then
  local yearlyArgs = {}
  for i = 2, yearlyCount * 2 + 1 do
    yearlyArgs[#yearlyArgs + 1] = ARGV[i]
  end
  redis.call('ZADD', KEYS[1], unpack(yearlyArgs))
end
local globalStart = yearlyCount * 2 + 2
if globalStart <= #ARGV then
  local globalArgs = {}
  for i = globalStart, #ARGV do
    globalArgs[#globalArgs + 1] = ARGV[i]
  end
  redis.call('ZADD', KEYS[2], unpack(globalArgs))
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

func (s *LeaderboardStore) UpdateContestScore(ctx context.Context, contestID uuid.UUID, userID uuid.UUID, score float64) (bool, error) {
	key := contestLeaderboardKey(contestID)
	scoreStr := strconv.FormatFloat(score, 'f', -1, 64)
	member := userID.String()

	result, err := updateScoreScript.Exec(ctx, s.client, []string{key}, []string{scoreStr, member}).ToInt64()
	if err != nil {
		return false, fmt.Errorf("failed to update contest leaderboard score for key %s: %w", key, err)
	}

	return result == 1, nil
}

func (s *LeaderboardStore) UpdateOfficialScores(ctx context.Context, year int, userID uuid.UUID, yearlyScore float64, globalScore float64) (bool, bool, error) {
	yearlyKey := yearlyLeaderboardKey(year)
	yearlyScoreStr := strconv.FormatFloat(yearlyScore, 'f', -1, 64)
	globalScoreStr := strconv.FormatFloat(globalScore, 'f', -1, 64)
	member := userID.String()

	result, err := updateOfficialScoresScript.Exec(ctx, s.client,
		[]string{yearlyKey, globalLeaderboardKey},
		[]string{yearlyScoreStr, globalScoreStr, member},
	).ToInt64()
	if err != nil {
		return false, false, fmt.Errorf("failed to update official leaderboard scores: %w", err)
	}

	yearlyUpdated := result&1 == 1
	globalUpdated := result&2 == 2
	return yearlyUpdated, globalUpdated, nil
}

func (s *LeaderboardStore) RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID, scores []domain.LeaderboardScore) error {
	key := contestLeaderboardKey(contestID)
	return s.rebuildLeaderboard(ctx, key, scores)
}

func (s *LeaderboardStore) RebuildOfficialLeaderboards(ctx context.Context, year int, yearlyScores []domain.LeaderboardScore, globalScores []domain.LeaderboardScore) error {
	yearlyKey := yearlyLeaderboardKey(year)

	// Build args: [yearlyCount, yearlyScore1, yearlyMember1, ..., globalScore1, globalMember1, ...]
	yearlyCount := len(yearlyScores)
	args := make([]string, 0, 1+yearlyCount*2+len(globalScores)*2)
	args = append(args, strconv.Itoa(yearlyCount))

	for _, entry := range yearlyScores {
		args = append(args, strconv.FormatFloat(entry.Score, 'f', -1, 64))
		args = append(args, entry.UserID.String())
	}
	for _, entry := range globalScores {
		args = append(args, strconv.FormatFloat(entry.Score, 'f', -1, 64))
		args = append(args, entry.UserID.String())
	}

	err := rebuildOfficialScript.Exec(ctx, s.client, []string{yearlyKey, globalLeaderboardKey}, args).Error()
	if err != nil {
		return fmt.Errorf("failed to rebuild official leaderboards: %w", err)
	}

	return nil
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
