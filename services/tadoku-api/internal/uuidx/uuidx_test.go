package uuidx_test

import (
	"bytes"
	"context"
	"math/rand"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/uuidx"
)

func TestNewReadsTheContextSource(t *testing.T) {
	want := []uuid.UUID{
		uuid.Must(uuid.NewRandomFromReader(bytes.NewReader(bytes.Repeat([]byte{0x11}, 16)))),
		uuid.Must(uuid.NewRandomFromReader(bytes.NewReader(bytes.Repeat([]byte{0x22}, 16)))),
	}
	ids := append(bytes.Repeat([]byte{0x11}, 16), bytes.Repeat([]byte{0x22}, 16)...)
	ctx := uuidx.WithSource(context.Background(), bytes.NewReader(ids))
	derived, cancel := context.WithCancel(context.WithoutCancel(ctx))
	defer cancel()

	got := []uuid.UUID{uuidx.New(ctx), uuidx.New(derived)}

	for index := range want {
		if got[index] != want[index] {
			t.Errorf("ID %d = %s, want %s", index, got[index], want[index])
		}
	}
}

func TestNewKeepsSourcesSeparate(t *testing.T) {
	first := uuidx.WithSource(context.Background(), rand.New(rand.NewSource(1)))
	second := uuidx.WithSource(context.Background(), rand.New(rand.NewSource(1)))

	a, b := uuidx.New(first), uuidx.New(second)

	if a != b {
		t.Errorf("same seed in separate contexts produced %s and %s", a, b)
	}
}

func TestNewWithoutSourceIsRandom(t *testing.T) {
	a, b := uuidx.New(context.Background()), uuidx.New(context.Background())

	if a == b || a.Version() != 4 || a.Variant() != uuid.RFC4122 {
		t.Errorf("IDs without a source = %s, %s; want distinct random version 4 IDs", a, b)
	}
}

func TestNewSynchronizesASharedSource(t *testing.T) {
	ctx := uuidx.WithSource(context.Background(), rand.New(rand.NewSource(1)))

	var mu sync.Mutex
	seen := make(map[uuid.UUID]bool)
	var callers sync.WaitGroup
	for range 16 {
		callers.Add(1)
		go func() {
			defer callers.Done()
			for range 100 {
				id := uuidx.New(ctx)
				mu.Lock()
				seen[id] = true
				mu.Unlock()
			}
		}()
	}
	callers.Wait()

	if len(seen) != 1600 {
		t.Errorf("concurrent callers produced %d distinct IDs, want 1600", len(seen))
	}
}

func TestNewV7SharesTheContextSource(t *testing.T) {
	ids := append(bytes.Repeat([]byte{0x11}, 16), bytes.Repeat([]byte{0x22}, 16)...)
	ctx := uuidx.WithSource(context.Background(), bytes.NewReader(ids))
	wantNext := uuid.Must(uuid.NewRandomFromReader(bytes.NewReader(bytes.Repeat([]byte{0x22}, 16))))

	ordered := uuidx.NewV7(ctx)
	next := uuidx.New(ctx)

	if ordered.Version() != 7 {
		t.Errorf("NewV7 version = %d, want 7", ordered.Version())
	}
	if next != wantNext {
		t.Errorf("ID after NewV7 = %s, want %s from the following 16 source bytes", next, wantNext)
	}
}

func TestNewV7WithoutSourceIsTimeOrdered(t *testing.T) {
	a, b := uuidx.NewV7(context.Background()), uuidx.NewV7(context.Background())

	if a == b || a.Version() != 7 || b.Version() != 7 {
		t.Errorf("IDs without a source = %s, %s; want distinct version 7 IDs", a, b)
	}
}
