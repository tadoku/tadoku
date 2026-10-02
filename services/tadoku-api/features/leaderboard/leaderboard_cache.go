package leaderboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	valkeygo "github.com/valkey-io/valkey-go"
)

const (
	contestPrefix = "leaderboard:contest:"
	yearlyPrefix  = "leaderboard:yearly:"
	globalKey     = "leaderboard:global"
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

type Cache struct {
	client           valkeygo.Client
	operationTimeout time.Duration
	cachePrefix      string
}

func NewCache(client valkeygo.Client, operationTimeout time.Duration, cachePrefix string) *Cache {
	return &Cache{
		client:           client,
		operationTimeout: operationTimeout,
		cachePrefix:      cachePrefix,
	}
}

func (c *Cache) cacheKey(key string) string { return c.cachePrefix + key }

func (c *Cache) fetchPage(ctx context.Context, key string, start int64, pageSize int) (*page, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.operationTimeout)
	defer cancel()
	marker, err := c.client.Do(ctx, c.client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err == valkeygo.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("check leaderboard marker %s: %w", key, err)
	}
	generation, err := c.generation(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if generation != "0" && marker != "native:"+generation {
		return nil, false, nil
	}
	total, err := c.client.Do(ctx, c.client.B().Zcard().Key(key).Build()).AsInt64()
	if err != nil {
		return nil, false, fmt.Errorf("get leaderboard cardinality %s: %w", key, err)
	}
	if total == 0 {
		current, err := c.cacheCurrent(ctx, key, marker, generation)
		if err != nil || !current {
			return nil, false, err
		}
		return &page{scores: []score{}, startRank: 1}, true, nil
	}

	stop := start + int64(pageSize) - 1
	fetchStart := start
	if fetchStart > 0 {
		fetchStart--
	}
	values, err := c.client.Do(ctx, c.client.B().Zrange().Key(key).Min(strconv.FormatInt(fetchStart, 10)).Max(strconv.FormatInt(stop+1, 10)).Rev().Withscores().Build()).AsZScores()
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
		higher, err := c.client.Do(ctx, c.client.B().Zcount().Key(key).Min("("+strconv.FormatFloat(scores[0].value, 'f', -1, 64)).Max("+inf").Build()).AsInt64()
		if err != nil {
			return nil, false, fmt.Errorf("count higher leaderboard scores %s: %w", key, err)
		}
		startRank = int(higher) + 1
	}
	current, err := c.cacheCurrent(ctx, key, marker, generation)
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

func (c *Cache) cacheCurrent(ctx context.Context, key, marker, generation string) (bool, error) {
	currentMarker, err := c.client.Do(ctx, c.client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err == valkeygo.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("recheck leaderboard marker %s: %w", key, err)
	}
	if currentMarker != marker {
		return false, nil
	}
	currentGeneration, err := c.generation(ctx, key)
	if err != nil {
		return false, err
	}
	return currentGeneration == generation, nil
}

func (c *Cache) generation(ctx context.Context, key string) (string, error) {
	value, err := c.client.Do(ctx, c.client.B().Get().Key(key+":generation").Build()).ToString()
	if err == valkeygo.Nil {
		return "0", nil
	}
	if err != nil {
		return "", fmt.Errorf("read leaderboard generation %s: %w", key, err)
	}
	return value, nil
}

func (c *Cache) invalidate(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, c.operationTimeout)
	defer cancel()
	if err := invalidateScript.Exec(ctx, c.client, []string{key, key + ":last_updated", key + ":generation"}, nil).Error(); err != nil {
		return fmt.Errorf("invalidate leaderboard %s: %w", key, err)
	}
	return nil
}

func (c *Cache) rebuild(ctx context.Context, key string, scores []score, generation string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.operationTimeout)
	defer cancel()
	args := make([]string, 0, 1+len(scores)*2)
	args = append(args, generation)
	for _, item := range scores {
		args = append(args, strconv.FormatFloat(item.value, 'f', -1, 64), item.userID.String())
	}
	published, err := rebuildScript.Exec(ctx, c.client, []string{key, key + ":last_updated", key + ":generation"}, args).ToInt64()
	if err != nil {
		return false, fmt.Errorf("rebuild leaderboard %s: %w", key, err)
	}
	return published == 1, nil
}
