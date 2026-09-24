package leaderboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestRebuildDoesNotPublishSnapshotAfterInvalidation(t *testing.T) {
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	option.SelectDB = 13
	option.ForceSingleClient = true
	option.DisableRetry = true
	client, err := valkeygo.NewClient(option)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)

	key := "leaderboard:global:test:" + uuid.NewString()
	lease := key + ":lease"
	if err := client.Do(t.Context(), client.B().Set().Key(lease).Value("test").Nx().Build()).Error(); err != nil {
		if errors.Is(err, valkeygo.Nil) {
			t.Fatal("test key is leased")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := client.Do(ctx, client.B().Del().Key(key, key+":last_updated", key+":generation", lease).Build()).Error(); err != nil {
			t.Error(err)
		}
	})

	service := NewService(nil, client, time.Second, "")
	before, err := service.generation(t.Context(), key)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.invalidate(t.Context(), key); err != nil {
		t.Fatal(err)
	}
	stale := []score{{userID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), value: 10}}
	published, err := service.rebuild(t.Context(), key, stale, before)
	if err != nil {
		t.Fatal(err)
	}
	if published {
		t.Error("stale snapshot was published after invalidation")
	}
	exists, err := client.Do(t.Context(), client.B().Exists().Key(key+":last_updated").Build()).AsInt64()
	if err != nil {
		t.Fatal(err)
	}
	if exists != 0 {
		t.Errorf("stale marker exists after rejected rebuild")
	}

	after, err := service.generation(t.Context(), key)
	if err != nil {
		t.Fatal(err)
	}
	fresh := []score{{userID: stale[0].userID, value: 20}}
	published, err = service.rebuild(t.Context(), key, fresh, after)
	if err != nil {
		t.Fatal(err)
	}
	if !published {
		t.Error("fresh snapshot was rejected")
	}
	marker, err := client.Do(t.Context(), client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err != nil {
		t.Fatal(err)
	}
	if marker != "native:"+after {
		t.Errorf("marker = %q, want native:%s", marker, after)
	}

	racing := NewService(nil, &invalidateBeforeZcard{Client: client, key: key, t: t}, time.Second, "")
	page, cacheExists, err := racing.fetchPage(t.Context(), key, 0, 25)
	if err != nil {
		t.Fatal(err)
	}
	if cacheExists || page != nil {
		t.Errorf("cache page after invalidation = %+v, exists = %t; want miss", page, cacheExists)
	}
}

type invalidateBeforeZcard struct {
	valkeygo.Client
	key   string
	t     *testing.T
	calls int
}

func (client *invalidateBeforeZcard) Do(ctx context.Context, command valkeygo.Completed) valkeygo.ValkeyResult {
	client.calls++
	if client.calls == 3 {
		if err := invalidateScript.Exec(ctx, client.Client, []string{client.key, client.key + ":last_updated", client.key + ":generation"}, nil).Error(); err != nil {
			client.t.Fatal(err)
		}
	}
	return client.Client.Do(ctx, command)
}
