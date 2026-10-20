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

var registrationAttempts = valkeygo.NewLuaScript(`
local ip = redis.call('INCR', KEYS[1])
if ip == 1 then redis.call('EXPIRE', KEYS[1], 900) end
if ip > 10 then return 0 end
return 1
`)

// admitRegistration fails closed: callers must not hash a password unless it returns nil.
func (t *Throttle) admitRegistration(ctx context.Context, ip string) error {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return errx.NewUnavailableError("tenant is required for registration admission", nil)
	}
	bounded, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	admitted, err := registrationAttempts.Exec(bounded, t.client, []string{
		"tenant:" + key.String() + ":identity:registration:ip:" + hashToken(ip),
	}, nil).AsInt64()
	if err != nil {
		return errx.NewUnavailableError("registration admission unavailable", err)
	}
	if admitted != 1 {
		return ErrRegistrationThrottled
	}
	return nil
}

var recoveryAdmissions = valkeygo.NewLuaScript(`
local client = redis.call('INCR', KEYS[1])
if client == 1 then redis.call('EXPIRE', KEYS[1], 900) end
if client > 10 then return 0 end
local recipient = redis.call('INCR', KEYS[2])
if recipient == 1 then redis.call('EXPIRE', KEYS[2], 900) end
if recipient > 3 then return 1 end
return 2
`)

// admitRecovery fails closed: callers must not write a token or mail unless it returns true.
// False without an error suppresses the recipient; answer exactly as for an unknown address.
func (t *Throttle) admitRecovery(ctx context.Context, email Email, ip string) (bool, error) {
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return false, errx.NewUnavailableError("tenant is required for recovery admission", nil)
	}
	prefix := "tenant:" + key.String() + ":identity:recovery:"
	bounded, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	admission, err := recoveryAdmissions.Exec(bounded, t.client, []string{
		prefix + "ip:" + hashToken(ip),
		prefix + "email:" + hashToken(email.value),
	}, nil).AsInt64()
	if err != nil {
		return false, errx.NewUnavailableError("recovery admission unavailable", err)
	}

	switch admission {
	case 0:
		return false, ErrRecoveryThrottled
	case 1:
		return false, nil
	}
	return true, nil
}

func hashToken(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}
