package valkey

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeaderboardStoreOperationTimeout(t *testing.T) {
	store := &LeaderboardStore{operationTimeout: time.Second}
	wantCtx, wantCancel := context.WithTimeout(context.Background(), time.Second)
	want, ok := wantCtx.Deadline()
	require.True(t, ok)
	wantCancel()

	ctx, cancel := store.withOperationTimeout(context.Background())
	defer cancel()
	got, ok := ctx.Deadline()
	require.True(t, ok)
	assert.WithinDuration(t, want, got, 100*time.Millisecond)
}

func TestLeaderboardStoreOperationTimeoutHonorsParentDeadline(t *testing.T) {
	store := &LeaderboardStore{operationTimeout: time.Minute}
	parent, cancelParent := context.WithTimeout(context.Background(), time.Second)
	defer cancelParent()

	ctx, cancel := store.withOperationTimeout(parent)
	defer cancel()

	want, ok := parent.Deadline()
	require.True(t, ok)
	got, ok := ctx.Deadline()
	require.True(t, ok)
	assert.Equal(t, want, got)
}
