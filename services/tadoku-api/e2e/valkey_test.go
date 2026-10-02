package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

const leaderboardValkeyLeaseKey = "tadoku-api:e2e:leaderboard:lease"

type leaderboardValkeyFixture struct {
	client valkeygo.Client
	token  string
}

func newLeaderboardValkeyFixture(ctx context.Context) (_ *leaderboardValkeyFixture, err error) {
	rawURL, err := testvalkey.URL()
	if err != nil {
		return nil, err
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse test Valkey URL: %w", err)
	}
	// Test safety: DB 15 is reserved for this E2E suite and DB 14 for its nested cleanup probe.
	option.SelectDB = leaderboardValkeyDatabase()
	option.ForceSingleClient = true
	option.DisableRetry = true
	client, err := valkeygo.NewClient(option)
	if err != nil {
		return nil, fmt.Errorf("open test Valkey: %w", err)
	}
	complete := false
	defer func() {
		if !complete {
			client.Close()
		}
	}()

	token := uuid.NewString()
	result := client.Do(ctx, client.B().Set().Key(leaderboardValkeyLeaseKey).Value(token).Nx().Build())
	if result.Error() != nil {
		if errors.Is(result.Error(), valkeygo.Nil) {
			return nil, fmt.Errorf("test Valkey DB %d is already leased", leaderboardValkeyDatabase())
		}
		return nil, fmt.Errorf("lease test Valkey DB %d: %w", leaderboardValkeyDatabase(), result.Error())
	}

	fixture := &leaderboardValkeyFixture{client: client, token: token}
	defer func() {
		if !complete {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = fixture.release(cleanupCtx)
		}
	}()

	keys, err := fixture.cacheKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("inspect test leaderboard keys: %w", err)
	}
	if len(keys) != 0 {
		return nil, fmt.Errorf("test Valkey DB %d contains leaderboard keys; refusing to overwrite them", leaderboardValkeyDatabase())
	}

	complete = true
	return fixture, nil
}

func newClosedLeaderboardValkeyClient() (valkeygo.Client, error) {
	rawURL, err := testvalkey.URL()
	if err != nil {
		return nil, err
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		return nil, err
	}
	option.SelectDB = leaderboardValkeyDatabase()
	client, err := valkeygo.NewClient(option)
	if err != nil {
		return nil, err
	}
	client.Close()
	return client, nil
}

func leaderboardValkeyDatabase() int {
	if os.Getenv("TADOKU_KRATOS_CLEANUP_FAILURE_TEST") != "" {
		return 14
	}
	return 15
}

func (f *leaderboardValkeyFixture) reset(ctx context.Context) error {
	if f == nil {
		return nil
	}
	if err := f.ownsLease(ctx); err != nil {
		return err
	}
	keys, err := f.cacheKeys(ctx)
	if err != nil || len(keys) == 0 {
		return err
	}
	return f.client.Do(ctx, f.client.B().Del().Key(keys...).Build()).Error()
}

func (f *leaderboardValkeyFixture) cacheKeys(ctx context.Context) ([]string, error) {
	var keys []string
	for _, pattern := range []string{"leaderboard:*", "tenant:*:leaderboard:*"} {
		var cursor uint64
		for {
			page, err := f.client.Do(
				ctx,
				f.client.B().Scan().Cursor(cursor).Match(pattern).Count(100).Build(),
			).AsScanEntry()
			if err != nil {
				return nil, err
			}
			keys = append(keys, page.Elements...)
			if page.Cursor == 0 {
				break
			}
			cursor = page.Cursor
		}
	}
	return keys, nil
}

func (f *leaderboardValkeyFixture) close() error {
	if f == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := f.reset(ctx)
	err = errors.Join(err, f.release(ctx))
	f.client.Close()
	return err
}

func (f *leaderboardValkeyFixture) release(ctx context.Context) error {
	if err := f.ownsLease(ctx); err != nil {
		return err
	}
	return f.client.Do(ctx, f.client.B().Del().Key(leaderboardValkeyLeaseKey).Build()).Error()
}

func (f *leaderboardValkeyFixture) ownsLease(ctx context.Context) error {
	token, err := f.client.Do(ctx, f.client.B().Get().Key(leaderboardValkeyLeaseKey).Build()).ToString()
	if err != nil {
		return fmt.Errorf("read test Valkey lease: %w", err)
	}
	if token != f.token {
		return fmt.Errorf("test Valkey DB %d lease ownership changed", leaderboardValkeyDatabase())
	}
	return nil
}
