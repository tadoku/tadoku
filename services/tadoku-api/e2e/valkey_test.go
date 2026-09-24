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

var leaderboardValkeyKeys = []string{
	"leaderboard:global",
	"leaderboard:global:last_updated",
	"leaderboard:global:generation",
	"leaderboard:yearly:2026",
	"leaderboard:yearly:2026:last_updated",
	"leaderboard:yearly:2026:generation",
	"leaderboard:yearly:2024",
	"leaderboard:yearly:2024:last_updated",
	"leaderboard:yearly:2024:generation",
	"leaderboard:contest:f1111111-1111-4111-8111-111111111111",
	"leaderboard:contest:f1111111-1111-4111-8111-111111111111:last_updated",
	"leaderboard:contest:f1111111-1111-4111-8111-111111111111:generation",
	"leaderboard:contest:f2222222-2222-4222-8222-222222222222",
	"leaderboard:contest:f2222222-2222-4222-8222-222222222222:last_updated",
	"leaderboard:contest:f2222222-2222-4222-8222-222222222222:generation",
	"leaderboard:contest:52fdfc07-2182-454f-963f-5f0f9a621d72",
	"leaderboard:contest:52fdfc07-2182-454f-963f-5f0f9a621d72:last_updated",
	"leaderboard:contest:52fdfc07-2182-454f-963f-5f0f9a621d72:generation",
	"leaderboard:contest:f0000000-0000-4000-8000-000000000001",
	"leaderboard:contest:f0000000-0000-4000-8000-000000000001:last_updated",
	"leaderboard:contest:f0000000-0000-4000-8000-000000000001:generation",
	"leaderboard:contest:f0000000-0000-4000-8000-000000000004",
	"leaderboard:contest:f0000000-0000-4000-8000-000000000004:last_updated",
	"leaderboard:contest:f0000000-0000-4000-8000-000000000004:generation",
}

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
	// A lease and an empty-key check make accidental overlap fail closed without
	// flushing or rewriting production keys.
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

	exists, err := client.Do(ctx, client.B().Exists().Key(leaderboardValkeyKeys...).Build()).ToInt64()
	if err != nil {
		return nil, fmt.Errorf("inspect test leaderboard keys: %w", err)
	}
	if exists != 0 {
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
	return f.client.Do(ctx, f.client.B().Del().Key(leaderboardValkeyKeys...).Build()).Error()
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
