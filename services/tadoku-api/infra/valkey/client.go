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

type dialContextFunc func(context.Context, string, *net.Dialer, *tls.Config) (net.Conn, error)

type startupState struct {
	ctx  context.Context
	done chan struct{}

	mu     sync.Mutex
	active bool
}

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

	startup := &startupState{ctx: ctx, done: make(chan struct{}), active: true}
	option.DialCtxFn = startupDialer(startup, option.DialCtxFn)
	client, err := valkeygo.NewClient(option)
	startup.finish()

	if ctxErr := ctx.Err(); ctxErr != nil {
		if client != nil {
			client.Close()
		}
		return nil, fmt.Errorf("open valkey: %w", ctxErr)
	}
	if client == nil {
		if err == nil {
			err = errors.New("client constructor returned nil")
		}
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
	if err != nil {
		return valkeygo.ClientOption{}, errors.New("invalid valkey URL")
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return valkeygo.ClientOption{}, errors.New("invalid valkey URL option")
	}
	allowedQuery := map[string]bool{
		"client_cache": true,
		"client_name":  true,
		"db":           true,
		"protocol":     true,
		"skip_verify":  true,
	}
	for key, values := range query {
		if !allowedQuery[key] || len(values) != 1 {
			return valkeygo.ClientOption{}, errors.New("invalid valkey URL option")
		}
	}
	if u.Fragment != "" || (u.Scheme != "unix" && u.Path != "" && query.Has("db")) {
		return valkeygo.ClientOption{}, errors.New("invalid valkey URL")
	}

	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		// ParseURL can include its input in URL syntax errors. Do not expose credentials.
		return valkeygo.ClientOption{}, errors.New("invalid valkey URL")
	}
	if len(option.InitAddress) != 1 || option.Sentinel.MasterSet != "" {
		return valkeygo.ClientOption{}, errors.New("valkey URL must configure exactly one standalone address")
	}
	if option.SelectDB < 0 {
		return valkeygo.ClientOption{}, errors.New("valkey database must not be negative")
	}
	if u.Scheme == "unix" {
		if option.InitAddress[0] == "" {
			return valkeygo.ClientOption{}, errors.New("valkey Unix socket path must not be empty")
		}
	} else {
		if u.Hostname() == "" {
			return valkeygo.ClientOption{}, errors.New("valkey URL must include a host and valid port")
		}
		host, port, err := net.SplitHostPort(option.InitAddress[0])
		if err != nil || host == "" {
			return valkeygo.ClientOption{}, errors.New("valkey URL must include a host and valid port")
		}
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return valkeygo.ClientOption{}, errors.New("valkey URL must include a host and valid port")
		}
	}

	option.Dialer.Timeout = timeout
	option.Dialer.KeepAlive = timeout
	option.ConnWriteTimeout = timeout
	option.ForceSingleClient = true
	option.DisableRetry = true
	option.AlwaysPipelining = true
	return option, nil
}

func (s *startupState) finish() {
	s.mu.Lock()
	s.active = false
	close(s.done)
	s.mu.Unlock()
}

func startupDialer(startup *startupState, parsed dialContextFunc) dialContextFunc {
	return func(ctx context.Context, address string, dialer *net.Dialer, tlsConfig *tls.Config) (net.Conn, error) {
		startup.mu.Lock()
		active := startup.active
		startup.mu.Unlock()
		if !active {
			return dial(ctx, address, dialer, tlsConfig, parsed)
		}

		dialContext, cancel := context.WithCancel(ctx)
		stop := context.AfterFunc(startup.ctx, cancel)
		connection, err := dial(dialContext, address, dialer, tlsConfig, parsed)
		stop()
		cancel()
		if err != nil {
			return nil, err
		}

		go func() {
			select {
			case <-startup.ctx.Done():
				startup.mu.Lock()
				if startup.active {
					_ = connection.Close()
				}
				startup.mu.Unlock()
			case <-startup.done:
			}
		}()
		return connection, nil
	}
}

func dial(ctx context.Context, address string, dialer *net.Dialer, tlsConfig *tls.Config, parsed dialContextFunc) (net.Conn, error) {
	if parsed != nil {
		return parsed(ctx, address, dialer, tlsConfig)
	}
	if tlsConfig != nil {
		return (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, "tcp", address)
	}
	return dialer.DialContext(ctx, "tcp", address)
}
