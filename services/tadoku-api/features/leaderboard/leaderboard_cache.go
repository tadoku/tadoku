package leaderboard

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	valkeygo "github.com/valkey-io/valkey-go"
)

const (
	contestPrefix = "leaderboard:contest:"
	yearlyPrefix  = "leaderboard:yearly:"
	globalKey     = "leaderboard:global"
)

var rebuildScript = valkeygo.NewLuaScript(`
if redis.call('GET', KEYS[3]) ~= ARGV[1] then
  return 0
end
redis.call('DEL', KEYS[2], KEYS[1])
for i = 2, #ARGV, 2000 do
  redis.call('ZADD', KEYS[1], unpack(ARGV, i, math.min(i + 1999, #ARGV)))
end
redis.call('SET', KEYS[2], 'native:' .. ARGV[1] .. ':' .. redis.call('ZCARD', KEYS[1]), 'EX', 2419200)
redis.call('EXPIRE', KEYS[3], 2419200)
return 1
`)

var pageScript = valkeygo.NewLuaScript(`
local generation = redis.call('GET', KEYS[3])
local size = redis.call('ZCARD', KEYS[1])
if not generation or redis.call('GET', KEYS[2]) ~= 'native:' .. generation .. ':' .. size then
  return false
end
local values = redis.call('ZRANGE', KEYS[1], ARGV[1], ARGV[2], 'REV', 'WITHSCORES')
local reply = {size, 0, 0}
if #values > 0 then
  reply[2] = redis.call('ZCOUNT', KEYS[1], '(' .. values[2], '+inf')
  reply[3] = redis.call('ZCOUNT', KEYS[1], values[#values], '+inf')
end
for _, value in ipairs(values) do
  table.insert(reply, value)
end
return reply
`)

var generationScript = valkeygo.NewLuaScript(`
local value = redis.call('GET', KEYS[1])
if not value or tonumber(value) < 4294967296 then
  redis.call('SET', KEYS[1], ARGV[1], 'EX', 3600)
  return ARGV[1]
end
return value
`)

var invalidateScript = valkeygo.NewLuaScript(`
if redis.call('EXISTS', KEYS[3]) == 1 then
  redis.call('INCR', KEYS[3])
else
  redis.call('SET', KEYS[3], ARGV[1], 'EX', 3600)
end
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
		return "", errx.NewUnavailableError("leaderboard cache requires a tenant", postgres.ErrNoTenant)
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

	bounds := []string{strconv.FormatInt(start, 10), strconv.FormatInt(start+int64(pageSize)-1, 10)}
	reply, err := pageScript.Exec(ctx, c.client, boardKeys(key), bounds).ToArray()
	if valkeygo.IsValkeyNil(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("fetch leaderboard page %s: %w", key, err)
	}

	board, err := parsePage(reply, int(start))
	if err != nil {
		return nil, false, fmt.Errorf("parse leaderboard page %s: %w", key, err)
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

func boardKeys(key string) []string {
	return []string{key, key + ":last_updated", key + ":generation"}
}

func newGeneration() string {
	return strconv.FormatInt(rand.Int64N(1<<62-1)+1, 10)
}

func (c *Cache) generation(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.operationTimeout)
	defer cancel()

	value, err := generationScript.Exec(
		ctx,
		c.client,
		[]string{key + ":generation"},
		[]string{newGeneration()},
	).ToString()
	if err != nil {
		return "", fmt.Errorf("read leaderboard generation %s: %w", key, err)
	}
	return value, nil
}

func (c *Cache) invalidate(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, c.operationTimeout)
	defer cancel()

	if err := invalidateScript.Exec(ctx, c.client, boardKeys(key), []string{newGeneration()}).Error(); err != nil {
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
	published, err := rebuildScript.Exec(ctx, c.client, boardKeys(key), args).ToInt64()
	if err != nil {
		return false, fmt.Errorf("rebuild leaderboard %s: %w", key, err)
	}
	return published == 1, nil
}
