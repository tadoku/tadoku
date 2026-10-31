package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	valkeygo "github.com/valkey-io/valkey-go"
)

type Throttle struct {
	client  valkeygo.Client
	timeout time.Duration
}

func NewThrottle(client valkeygo.Client, timeout time.Duration) *Throttle {
	return &Throttle{client: client, timeout: timeout}
}

var errCountersUnavailable = errors.New("identity throttle counters unavailable")

type loginCounts struct {
	email int64
	ip    int64
}

type recoveryCounts struct {
	client    int64
	recipient int64
}

var chargeCounters = valkeygo.NewLuaScript(`
local counts = {}
for i, key in ipairs(KEYS) do
	counts[i] = redis.call('INCR', key)
	if counts[i] == 1 then redis.call('EXPIRE', key, ARGV[1]) end
end
return counts
`)

var chargeRecovery = valkeygo.NewLuaScript(`
local client = redis.call('INCR', KEYS[1])
if client == 1 then redis.call('EXPIRE', KEYS[1], ARGV[1]) end
if client > tonumber(ARGV[2]) then return {client, 0} end
local recipient = redis.call('INCR', KEYS[2])
if recipient == 1 then redis.call('EXPIRE', KEYS[2], ARGV[1]) end
return {client, recipient}
`)

func (t *Throttle) countLogin(ctx context.Context, window time.Duration, email Email, ip string) (loginCounts, error) {
	prefix, err := throttlePrefix(ctx, "login")
	if err != nil {
		return loginCounts{}, err
	}

	keys := []string{prefix + "email:" + hashToken(email.value), prefix + "ip:" + hashToken(ip)}
	counts, err := t.charge(ctx, chargeCounters, keys, window)
	if err != nil {
		return loginCounts{}, err
	}
	return loginCounts{email: counts[0], ip: counts[1]}, nil
}

func (t *Throttle) countRegistration(ctx context.Context, window time.Duration, ip string) (int64, error) {
	prefix, err := throttlePrefix(ctx, "registration")
	if err != nil {
		return 0, err
	}

	counts, err := t.charge(ctx, chargeCounters, []string{prefix + "ip:" + hashToken(ip)}, window)
	if err != nil {
		return 0, err
	}
	return counts[0], nil
}

// countRecovery charges the recipient only while the client count stays within clientLimit;
// otherwise the recipient count is zero.
func (t *Throttle) countRecovery(
	ctx context.Context,
	window time.Duration,
	clientLimit int64,
	email Email,
	ip string,
) (recoveryCounts, error) {
	prefix, err := throttlePrefix(ctx, "recovery")
	if err != nil {
		return recoveryCounts{}, err
	}

	keys := []string{prefix + "ip:" + hashToken(ip), prefix + "email:" + hashToken(email.value)}
	counts, err := t.charge(ctx, chargeRecovery, keys, window, strconv.FormatInt(clientLimit, 10))
	if err != nil {
		return recoveryCounts{}, err
	}
	return recoveryCounts{client: counts[0], recipient: counts[1]}, nil
}

func throttlePrefix(ctx context.Context, operation string) (string, error) {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return "", errors.New("tenant is required for identity throttling")
	}
	return "tenant:" + key.String() + ":identity:" + operation + ":", nil
}

func (t *Throttle) charge(
	ctx context.Context,
	script *valkeygo.Lua,
	keys []string,
	window time.Duration,
	args ...string,
) ([]int64, error) {
	bounded, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	args = append([]string{strconv.FormatInt(int64(window/time.Second), 10)}, args...)
	counts, err := script.Exec(bounded, t.client, keys, args).AsIntSlice()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errCountersUnavailable, err)
	}
	return counts, nil
}

func hashToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}
