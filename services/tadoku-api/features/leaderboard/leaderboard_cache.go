package leaderboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
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

var pageScript = valkeygo.NewLuaScript(`
local values = redis.call('ZRANGE', KEYS[1], ARGV[1], ARGV[2], 'REV', 'WITHSCORES')
local reply = {redis.call('ZCARD', KEYS[1]), 0, 0}
if #values > 0 then
  reply[2] = redis.call('ZCOUNT', KEYS[1], '(' .. values[2], '+inf')
  reply[3] = redis.call('ZCOUNT', KEYS[1], values[#values], '+inf')
end
for _, value in ipairs(values) do
  table.insert(reply, value)
end
return reply
`)

var invalidateScript = valkeygo.NewLuaScript(`
redis.call('INCR', KEYS[3])
redis.call('DEL', KEYS[1], KEYS[2])
return 1
`)

type Cache struct {
	client           valkeygo.Client
	operationTimeout time.Duration
}

func NewCache(client valkeygo.Client, operationTimeout time.Duration) *Cache {
	return &Cache{
		client:           client,
		operationTimeout: operationTimeout,
	}
}

func (c *Cache) cacheKey(ctx context.Context, key string) (string, error) {
	tenantKey, ok := tenant.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("leaderboard cache requires a tenant")
	}

	return "tenant:" + tenantKey.String() + ":" + key, nil
}

func (c *Cache) DeleteTestTenant(ctx context.Context, key tenant.TestKey) error {
	if key.String() == "" {
		return fmt.Errorf("leaderboard deletion requires a test tenant")
	}
	prefix, err := c.cacheKey(tenant.WithKey(ctx, key.Key()), "")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, c.operationTimeout)
	defer cancel()
	var cursor uint64
	for {
		page, err := c.client.Do(ctx, c.client.B().Scan().Cursor(cursor).Match(prefix+"*").Count(500).Build()).AsScanEntry()
		if err != nil {
			return fmt.Errorf("scan tenant leaderboard keys: %w", err)
		}

		for start := 0; start < len(page.Elements); start += 500 {
			end := min(start+500, len(page.Elements))
			err := c.client.Do(ctx, c.client.B().Unlink().Key(page.Elements[start:end]...).Build()).Error()
			if err != nil {
				return fmt.Errorf("unlink tenant leaderboard keys: %w", err)
			}
		}
		if page.Cursor == 0 {
			return nil
		}
		cursor = page.Cursor
	}
}

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

	bounds := []string{strconv.FormatInt(start, 10), strconv.FormatInt(start+int64(pageSize)-1, 10)}
	reply, err := pageScript.Exec(ctx, c.client, []string{key}, bounds).ToArray()
	if err != nil {
		return nil, false, fmt.Errorf("fetch leaderboard page %s: %w", key, err)
	}
	board, err := parsePage(reply, int(start))
	if err != nil {
		return nil, false, fmt.Errorf("parse leaderboard page %s: %w", key, err)
	}

	current, err := c.cacheCurrent(ctx, key, marker, generation)
	if err != nil || !current {
		return nil, false, err
	}
	return board, true, nil
}

func parsePage(reply []valkeygo.ValkeyMessage, offset int) (*page, error) {
	counts := make([]int64, 3)
	for i := range counts {
		count, err := reply[i].AsInt64()
		if err != nil {
			return nil, err
		}
		counts[i] = count
	}

	members := reply[3:]
	board := &page{
		scores:      make([]score, len(members)/2),
		offset:      offset,
		totalCount:  int(counts[0]),
		aboveFirst:  int(counts[1]),
		throughLast: int(counts[2]),
	}
	for i := range board.scores {
		member, err := members[2*i].ToString()
		if err != nil {
			return nil, err
		}
		id, err := uuid.Parse(member)
		if err != nil {
			return nil, fmt.Errorf("parse member %q: %w", member, err)
		}
		value, err := members[2*i+1].AsFloat64()
		if err != nil {
			return nil, err
		}
		board.scores[i] = score{userID: id, value: value}
	}
	return board, nil
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
	ctx, cancel := context.WithTimeout(ctx, c.operationTimeout)
	defer cancel()
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
