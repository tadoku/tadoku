package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	valkeygo "github.com/valkey-io/valkey-go"
)

type Throttle struct {
	client  valkeygo.Client
	timeout time.Duration
	logger  *slog.Logger
}

func NewThrottle(client valkeygo.Client, timeout time.Duration, logger *slog.Logger) *Throttle {
	return &Throttle{client: client, timeout: timeout, logger: logger}
}

var loginAttempts = valkeygo.NewLuaScript(`
local email = redis.call('INCR', KEYS[1])
if email == 1 then redis.call('EXPIRE', KEYS[1], 900) end
local ip = redis.call('INCR', KEYS[2])
if ip == 1 then redis.call('EXPIRE', KEYS[2], 900) end
if email > 10 or ip > 50 then return 0 end
return 1
`)

func (t *Throttle) allow(ctx context.Context, email Email, ip string) error {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return errx.NewUnavailableError("tenant is required for login throttling", nil)
	}
	prefix := "tenant:" + key.String() + ":identity:login:"
	bounded, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	allowed, err := loginAttempts.Exec(bounded, t.client, []string{
		prefix + "email:" + hashToken(email.value),
		prefix + "ip:" + hashToken(ip),
	}, nil).AsInt64()
	if err != nil {
		t.logger.WarnContext(ctx, "login throttling unavailable; allowing attempt", "error", err)
		return nil
	}
	if allowed != 1 {
		return ErrThrottled
	}
	return nil
}

func hashToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}
