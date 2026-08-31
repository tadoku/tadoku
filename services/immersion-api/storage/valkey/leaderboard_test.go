package valkey

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	valkeylib "github.com/valkey-io/valkey-go"
)

type contextCapturingClient struct {
	valkeylib.Client
	contexts []context.Context
}

func (c *contextCapturingClient) Do(ctx context.Context, cmd valkeylib.Completed) valkeylib.ValkeyResult {
	c.contexts = append(c.contexts, ctx)
	return c.Client.Do(ctx, cmd)
}

func newUnavailableContextCapturingClient(t *testing.T) *contextCapturingClient {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	client, err := valkeylib.NewClient(valkeylib.ClientOption{
		InitAddress:       []string{address},
		ForceSingleClient: true,
		DisableRetry:      true,
		Dialer: net.Dialer{
			Timeout: 10 * time.Millisecond,
		},
	})
	require.Error(t, err)
	require.NotNil(t, client)
	t.Cleanup(client.Close)

	return &contextCapturingClient{Client: client}
}

func TestLeaderboardStorePublicOperationsApplyOperationTimeout(t *testing.T) {
	client := newUnavailableContextCapturingClient(t)
	operationTimeout := time.Second
	store := NewLeaderboardStore(
		client,
		commondomain.NewMockClock(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)),
		operationTimeout,
	)
	contestID := uuid.New()
	userID := uuid.New()
	scores := []domain.LeaderboardScore{{UserID: userID, Score: 10}}

	tests := []struct {
		name string
		run  func(context.Context)
	}{
		{
			name: "update contest score",
			run: func(ctx context.Context) {
				_, _ = store.UpdateContestScore(ctx, contestID, userID, 10)
			},
		},
		{
			name: "update official scores",
			run: func(ctx context.Context) {
				_, _, _ = store.UpdateOfficialScores(ctx, 2026, userID, 10, 20)
			},
		},
		{
			name: "rebuild contest leaderboard",
			run: func(ctx context.Context) {
				_ = store.RebuildContestLeaderboard(ctx, contestID, scores)
			},
		},
		{
			name: "rebuild official leaderboards",
			run: func(ctx context.Context) {
				_ = store.RebuildOfficialLeaderboards(ctx, 2026, scores, scores)
			},
		},
		{
			name: "fetch global leaderboard",
			run: func(ctx context.Context) {
				_, _, _ = store.FetchGlobalLeaderboardPage(ctx, 0, 20)
			},
		},
		{
			name: "fetch yearly leaderboard",
			run: func(ctx context.Context) {
				_, _, _ = store.FetchYearlyLeaderboardPage(ctx, 2026, 0, 20)
			},
		},
		{
			name: "fetch contest leaderboard",
			run: func(ctx context.Context) {
				_, _, _ = store.FetchContestLeaderboardPage(ctx, contestID, 0, 20)
			},
		},
		{
			name: "rebuild global leaderboard",
			run: func(ctx context.Context) {
				_ = store.RebuildGlobalLeaderboard(ctx, scores)
			},
		},
		{
			name: "rebuild yearly leaderboard",
			run: func(ctx context.Context) {
				_ = store.RebuildYearlyLeaderboard(ctx, 2026, scores)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client.contexts = nil
			expectedCtx, expectedCancel := context.WithTimeout(context.Background(), operationTimeout)
			expectedDeadline, ok := expectedCtx.Deadline()
			require.True(t, ok)
			expectedCancel()

			test.run(context.Background())

			require.NotEmpty(t, client.contexts)
			deadline, ok := client.contexts[0].Deadline()
			require.True(t, ok)
			assert.WithinDuration(t, expectedDeadline, deadline, 100*time.Millisecond)
		})
	}
}

func TestLeaderboardStoreOperationTimeoutHonorsShorterParentDeadline(t *testing.T) {
	client := newUnavailableContextCapturingClient(t)
	store := NewLeaderboardStore(
		client,
		commondomain.NewMockClock(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)),
		time.Minute,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	parentDeadline, ok := ctx.Deadline()
	require.True(t, ok)

	_, _ = store.UpdateContestScore(ctx, uuid.New(), uuid.New(), 10)

	require.NotEmpty(t, client.contexts)
	operationDeadline, ok := client.contexts[0].Deadline()
	require.True(t, ok)
	assert.Equal(t, parentDeadline, operationDeadline)
}
