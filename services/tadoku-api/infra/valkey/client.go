// Package valkey constructs the raw Valkey client owned by tadoku-api.
package valkey

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	valkeygo "github.com/valkey-io/valkey-go"
)

// Open constructs a standalone client. A non-nil client returned with an error
// can reconnect on a later command and must still be closed by the caller.
func Open(ctx context.Context, rawURL string, timeout time.Duration) (valkeygo.Client, error) {
	option, err := clientOption(rawURL, timeout)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("open valkey: %w", err)
	}

	// NewClient has no context, so bridge ctx through its initial dial and
	// handshake. Detaching under the mutex leaves later reconnect contexts intact.
	var mu sync.Mutex
	initializing := true
	var initialConnection net.Conn
	stopClose := context.AfterFunc(ctx, func() {
		mu.Lock()
		defer mu.Unlock()
		if initializing && initialConnection != nil {
			_ = initialConnection.Close()
		}
	})
	option.DialCtxFn = func(dialContext context.Context, address string, dialer *net.Dialer, tlsConfig *tls.Config) (net.Conn, error) {
		mu.Lock()
		initial := initializing
		mu.Unlock()
		if initial {
			dialContext = ctx
		}

		var connection net.Conn
		var dialErr error
		if tlsConfig != nil {
			connection, dialErr = (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(dialContext, "tcp", address)
		} else {
			connection, dialErr = dialer.DialContext(dialContext, "tcp", address)
		}
		if dialErr != nil {
			return nil, dialErr
		}
		if initial {
			mu.Lock()
			if initializing && ctx.Err() == nil {
				initialConnection = connection
			} else {
				_ = connection.Close()
				dialErr = ctx.Err()
			}
			mu.Unlock()
		}
		return connection, dialErr
	}

	client, err := valkeygo.NewClient(option)
	mu.Lock()
	initializing = false
	initialConnection = nil
	mu.Unlock()
	stopClose()

	if ctxErr := ctx.Err(); ctxErr != nil {
		if client != nil {
			client.Close()
		}
		return nil, fmt.Errorf("open valkey: %w", ctxErr)
	}
	if client == nil && err != nil {
		return nil, fmt.Errorf("open valkey: %w", err)
	}
	if err != nil {
		return client, fmt.Errorf("connect to valkey: %w", err)
	}
	return client, nil
}

func clientOption(rawURL string, timeout time.Duration) (valkeygo.ClientOption, error) {
	if timeout <= 0 {
		return valkeygo.ClientOption{}, errors.New("valkey timeout must be positive")
	}
	u, err := url.Parse(rawURL)
	if err != nil ||
		(u.Scheme != "redis" && u.Scheme != "rediss") ||
		u.Hostname() == "" ||
		u.Path != "" ||
		u.ForceQuery ||
		u.RawQuery != "" ||
		u.Fragment != "" {
		return valkeygo.ClientOption{}, errors.New("invalid valkey URL")
	}

	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		return valkeygo.ClientOption{}, errors.New("invalid valkey URL")
	}
	_, port, err := net.SplitHostPort(option.InitAddress[0])
	if err != nil {
		return valkeygo.ClientOption{}, errors.New("invalid valkey URL")
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return valkeygo.ClientOption{}, errors.New("invalid valkey URL")
	}

	option.Dialer.Timeout = timeout
	option.Dialer.KeepAlive = timeout
	option.ConnWriteTimeout = timeout
	option.ForceSingleClient = true
	option.DisableRetry = true
	option.AlwaysPipelining = true
	return option, nil
}
