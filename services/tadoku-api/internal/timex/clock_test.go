package timex_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestClockSetAndAdvance(t *testing.T) {
	instant := time.Date(2026, time.September, 12, 21, 0, 0, 123456789, time.FixedZone("JST", 9*60*60))
	clock := timex.NewClock(instant)

	if got := clock.Now(); got != instant.UTC() {
		t.Errorf("new clock Now() = %#v, want %#v", got, instant.UTC())
	}

	clock.Advance(90 * time.Minute)
	if got, want := clock.Now(), instant.UTC().Add(90*time.Minute); got != want {
		t.Errorf("advanced Now() = %#v, want %#v", got, want)
	}

	next := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.FixedZone("offset", -5*60*60))
	clock.Set(next)
	if got := clock.Now(); got != next.UTC() {
		t.Errorf("set Now() = %#v, want %#v", got, next.UTC())
	}
}

func TestClockConcurrentAdvance(t *testing.T) {
	instant := time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC)
	clock := timex.NewClock(instant)

	var writers sync.WaitGroup
	for range 32 {
		writers.Add(1)
		go func() {
			defer writers.Done()
			for range 100 {
				clock.Advance(time.Second)
				_ = clock.Now()
			}
		}()
	}
	writers.Wait()

	if got, want := clock.Now(), instant.Add(3200*time.Second); got != want {
		t.Errorf("Now() after concurrent advances = %#v, want %#v", got, want)
	}
}

func TestNowReadsTheContextClock(t *testing.T) {
	first := timex.NewClock(time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC))
	second := timex.NewClock(time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC))
	firstCtx := timex.WithClock(context.Background(), first)
	secondCtx := timex.WithClock(context.Background(), second)
	derived, cancel := context.WithTimeout(context.WithoutCancel(firstCtx), time.Minute)
	defer cancel()

	first.Advance(time.Hour)

	if got, want := timex.Now(firstCtx), first.Now(); got != want {
		t.Errorf("Now(first) = %#v, want %#v", got, want)
	}
	if got, want := timex.Now(derived), first.Now(); got != want {
		t.Errorf("Now(derived from first) = %#v, want %#v", got, want)
	}
	if got, want := timex.Now(secondCtx), second.Now(); got != want {
		t.Errorf("Now(second) = %#v, want %#v", got, want)
	}
}

func TestNowPrefersTheContextClockOverTheWorld(t *testing.T) {
	clock := timex.NewClock(time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC))
	world := time.Date(1999, time.December, 31, 23, 59, 59, 0, time.UTC)

	timex.TheWorld(world, func() {
		if got, want := timex.Now(timex.WithClock(context.Background(), clock)), clock.Now(); got != want {
			t.Errorf("Now(with clock) = %#v, want %#v", got, want)
		}
		if got := timex.Now(context.Background()); got != world {
			t.Errorf("Now(without clock) = %#v, want %#v", got, world)
		}
	})
}
