package leaderboard

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	valkeygo "github.com/valkey-io/valkey-go"
)

const (
	contestPrefix = "leaderboard:contest:"
	yearlyPrefix  = "leaderboard:yearly:"
	globalKey     = "leaderboard:global"
	readinessKey  = "leaderboard:ready"
)

var rebuildScript = valkeygo.NewLuaScript(`
if (redis.call('GET', KEYS[3]) or '0') ~= ARGV[1] then
  return 0
end
redis.call('DEL', KEYS[1])
if #ARGV > 1 then
  redis.call('ZADD', KEYS[1], unpack(ARGV, 2))
end
redis.call('SET', KEYS[2], 'native:' .. ARGV[1])
return 1
`)

var invalidateScript = valkeygo.NewLuaScript(`
redis.call('INCR', KEYS[3])
redis.call('DEL', KEYS[1], KEYS[2])
return 1
`)

type Store struct {
	client           valkeygo.Client
	operationTimeout time.Duration
	cachePrefix      string
}

func NewStore(client valkeygo.Client, operationTimeout time.Duration, cachePrefix string) *Store {
	return &Store{
		client:           client,
		operationTimeout: operationTimeout,
		cachePrefix:      cachePrefix,
	}
}

func (s *Store) cacheKey(key string) string { return s.cachePrefix + key }

func (s *Store) readiness(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	count, err := s.client.Do(ctx, s.client.B().Exists().Key(s.cacheKey(readinessKey)).Build()).AsInt64()
	if err != nil {
		return false, fmt.Errorf("check leaderboard cache readiness: %w", err)
	}
	return count == 1, nil
}

func (s *Store) publishReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	if err := s.client.Do(ctx, s.client.B().Arbitrary("SET", s.cacheKey(readinessKey), "1", "EX", "5").Build()).Error(); err != nil {
		return fmt.Errorf("publish leaderboard cache readiness: %w", err)
	}
	return nil
}

func (s *Store) revokeReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	if err := s.client.Do(ctx, s.client.B().Del().Key(s.cacheKey(readinessKey)).Build()).Error(); err != nil {
		return fmt.Errorf("revoke leaderboard cache readiness: %w", err)
	}
	return nil
}

func (s *Store) fetchPage(ctx context.Context, key string, currentPage, pageSize int) (*page, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	marker, err := s.client.Do(ctx, s.client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err == valkeygo.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("check leaderboard marker %s: %w", key, err)
	}
	generation, err := s.generation(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if generation != "0" && marker != "native:"+generation {
		return nil, false, nil
	}
	total, err := s.client.Do(ctx, s.client.B().Zcard().Key(key).Build()).AsInt64()
	if err != nil {
		return nil, false, fmt.Errorf("get leaderboard cardinality %s: %w", key, err)
	}
	if total == 0 {
		current, err := s.cacheCurrent(ctx, key, marker, generation)
		if err != nil || !current {
			return nil, false, err
		}
		return &page{scores: []score{}, startRank: 1}, true, nil
	}

	start := int64(currentPage * pageSize)
	stop := start + int64(pageSize) - 1
	fetchStart := start
	if fetchStart > 0 {
		fetchStart--
	}
	values, err := s.client.Do(ctx, s.client.B().Zrange().Key(key).Min(strconv.FormatInt(fetchStart, 10)).Max(strconv.FormatInt(stop+1, 10)).Rev().Withscores().Build()).AsZScores()
	if err != nil {
		return nil, false, fmt.Errorf("fetch leaderboard page %s: %w", key, err)
	}
	hasPrevious := fetchStart < start && len(values) > 0
	previous := float64(0)
	if hasPrevious {
		previous = values[0].Score
		values = values[1:]
	}
	hasNext := int64(len(values)) > int64(pageSize)
	next := float64(0)
	if hasNext {
		next = values[len(values)-1].Score
		values = values[:len(values)-1]
	}
	scores := make([]score, len(values))
	for i, value := range values {
		id, err := uuid.Parse(value.Member)
		if err != nil {
			return nil, false, fmt.Errorf("parse leaderboard member %q: %w", value.Member, err)
		}
		scores[i] = score{userID: id, value: value.Score}
	}
	startRank := int(start) + 1
	if len(scores) > 0 {
		higher, err := s.client.Do(ctx, s.client.B().Zcount().Key(key).Min("("+strconv.FormatFloat(scores[0].value, 'f', -1, 64)).Max("+inf").Build()).AsInt64()
		if err != nil {
			return nil, false, fmt.Errorf("count higher leaderboard scores %s: %w", key, err)
		}
		startRank = int(higher) + 1
	}
	current, err := s.cacheCurrent(ctx, key, marker, generation)
	if err != nil || !current {
		return nil, false, err
	}
	return &page{
		scores:     scores,
		totalCount: int(total),
		startRank:  startRank,
		hasPrevTie: hasPrevious && len(scores) > 0 && previous == scores[0].value,
		hasNextTie: hasNext && len(scores) > 0 && next == scores[len(scores)-1].value,
	}, true, nil
}

func (s *Store) cacheCurrent(ctx context.Context, key, marker, generation string) (bool, error) {
	currentMarker, err := s.client.Do(ctx, s.client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err == valkeygo.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("recheck leaderboard marker %s: %w", key, err)
	}
	if currentMarker != marker {
		return false, nil
	}
	currentGeneration, err := s.generation(ctx, key)
	if err != nil {
		return false, err
	}
	return currentGeneration == generation, nil
}

func (s *Store) generation(ctx context.Context, key string) (string, error) {
	value, err := s.client.Do(ctx, s.client.B().Get().Key(key+":generation").Build()).ToString()
	if err == valkeygo.Nil {
		return "0", nil
	}
	if err != nil {
		return "", fmt.Errorf("read leaderboard generation %s: %w", key, err)
	}
	return value, nil
}

func (s *Store) invalidate(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	if err := invalidateScript.Exec(ctx, s.client, []string{key, key + ":last_updated", key + ":generation"}, nil).Error(); err != nil {
		return fmt.Errorf("invalidate leaderboard %s: %w", key, err)
	}
	return nil
}

func (s *Store) reconcile(ctx context.Context) (int, error) {
	var cursor uint64
	count := 0
	for {
		commandCtx, cancel := context.WithTimeout(ctx, s.operationTimeout)
		page, err := s.client.Do(commandCtx, s.client.B().Scan().Cursor(cursor).Match(s.cacheKey("leaderboard:*:last_updated")).Count(100).Build()).AsScanEntry()
		cancel()
		if err != nil {
			return count, fmt.Errorf("scan leaderboard cache markers: %w", err)
		}
		for _, marker := range page.Elements {
			key := strings.TrimSuffix(marker, ":last_updated")
			if !s.leaderboardCacheKey(key) {
				continue
			}
			if err := s.invalidate(ctx, key); err != nil {
				return count, err
			}
			count++
		}
		if page.Cursor == 0 {
			return count, nil
		}
		cursor = page.Cursor
	}
}

func (s *Store) leaderboardCacheKey(key string) bool {
	if !strings.HasPrefix(key, s.cachePrefix) {
		return false
	}
	key = strings.TrimPrefix(key, s.cachePrefix)
	if key == globalKey {
		return true
	}
	if strings.HasPrefix(key, yearlyPrefix) {
		_, err := strconv.Atoi(strings.TrimPrefix(key, yearlyPrefix))
		return err == nil
	}
	if strings.HasPrefix(key, contestPrefix) {
		_, err := uuid.Parse(strings.TrimPrefix(key, contestPrefix))
		return err == nil
	}
	return false
}

func (s *Store) rebuild(ctx context.Context, key string, scores []score, generation string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	args := make([]string, 0, 1+len(scores)*2)
	args = append(args, generation)
	for _, item := range scores {
		args = append(args, strconv.FormatFloat(item.value, 'f', -1, 64), item.userID.String())
	}
	published, err := rebuildScript.Exec(ctx, s.client, []string{key, key + ":last_updated", key + ":generation"}, args).ToInt64()
	if err != nil {
		return false, fmt.Errorf("rebuild leaderboard %s: %w", key, err)
	}
	return published == 1, nil
}
