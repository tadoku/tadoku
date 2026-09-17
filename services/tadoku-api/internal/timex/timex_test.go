package timex_test

import (
	"sync"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex/timextest"
)

func TestNow(t *testing.T) {
	before := time.Now().UTC().Truncate(time.Microsecond)
	got := timex.Now()
	after := time.Now().UTC().Truncate(time.Microsecond)
	if got.Location() != time.UTC || got.Before(before) || got.After(after) {
		t.Errorf("Now() = %v (%v), want UTC time between %v and %v", got, got.Location(), before, after)
	}
	if got.Nanosecond()%1000 != 0 {
		t.Errorf("Now() nanoseconds = %d, want microsecond precision", got.Nanosecond())
	}
}

func TestTheWorldFixedUTCAndRestoration(t *testing.T) {
	for _, instant := range []time.Time{
		time.Date(1998, time.March, 4, 5, 6, 7, 123456789, time.FixedZone("offset", -7*60*60)),
		{},
		time.Date(2044, time.July, 12, 10, 11, 12, 13, time.UTC),
	} {
		called := false
		timextest.TheWorld(instant, func() {
			called = true
			for i := 0; i < 3; i++ {
				if got := timex.Now(); got != instant.UTC() {
					t.Errorf("Now() = %#v, want %#v", got, instant.UTC())
				}
			}
		})
		if !called {
			t.Error("TheWorld did not call its callback")
		}
		before := time.Now().UTC().Truncate(time.Microsecond)
		got := timex.Now()
		after := time.Now().UTC().Truncate(time.Microsecond)
		if got.Location() != time.UTC || got.Before(before) || got.After(after) {
			t.Errorf("after scope: Now() = %v (%v), want UTC time between %v and %v", got, got.Location(), before, after)
		}
	}
}

func TestTheWorldConcurrentReaders(t *testing.T) {
	instant := time.Date(2001, time.February, 3, 4, 5, 6, 7, time.UTC)
	timextest.TheWorld(instant, func() {
		var readers sync.WaitGroup
		for i := 0; i < 32; i++ {
			readers.Add(1)
			go func() {
				defer readers.Done()
				for j := 0; j < 1000; j++ {
					if got := timex.Now(); got != instant {
						t.Errorf("concurrent Now() = %#v, want %#v", got, instant)
						return
					}
				}
			}()
		}
		readers.Wait()
	})
}

func TestTheWorldRestoresAfterPanic(t *testing.T) {
	panicValue := &struct{ reason string }{"callback failed"}
	func() {
		defer func() {
			if got := recover(); got != panicValue {
				t.Errorf("panic = %#v, want original value %#v", got, panicValue)
			}
		}()
		timextest.TheWorld(time.Time{}, func() { panic(panicValue) })
	}()
	before := time.Now().UTC().Truncate(time.Microsecond)
	got := timex.Now()
	after := time.Now().UTC().Truncate(time.Microsecond)
	if got.Location() != time.UTC || got.Before(before) || got.After(after) {
		t.Errorf("after panic: Now() = %v (%v), want UTC time between %v and %v", got, got.Location(), before, after)
	}
	instant := time.Date(2050, time.January, 2, 3, 4, 5, 6, time.UTC)
	timextest.TheWorld(instant, func() {
		if got := timex.Now(); got != instant {
			t.Errorf("later scope: Now() = %#v, want %#v", got, instant)
		}
	})
}

func TestTheWorldRejectsNestedOverride(t *testing.T) {
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		timextest.TheWorld(time.Time{}, func() {
			timextest.TheWorld(time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC), func() {
				panic("nested callback must not run")
			})
		})
	}()
	select {
	case got := <-done:
		if got != "timextest.TheWorld: override already active" {
			t.Fatalf("nested override panic = %#v, want clear rejection", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("nested override deadlocked")
	}
	timextest.TheWorld(time.Time{}, func() {
		if got := timex.Now(); got != (time.Time{}) {
			t.Errorf("scope after nested rejection: Now() = %#v, want zero time", got)
		}
	})
}

func TestTheWorldRejectsConcurrentOverride(t *testing.T) {
	instant := time.Date(2020, time.April, 5, 6, 7, 8, 9, time.UTC)
	timextest.TheWorld(instant, func() {
		done := make(chan any, 1)
		go func() {
			defer func() { done <- recover() }()
			timextest.TheWorld(time.Time{}, func() {
				panic("concurrent callback must not run")
			})
		}()
		select {
		case got := <-done:
			if got != "timextest.TheWorld: override already active" {
				t.Errorf("concurrent override panic = %#v, want clear rejection", got)
			}
		case <-time.After(5 * time.Second):
			t.Error("concurrent override blocked behind the active scope")
		}
		if got := timex.Now(); got != instant {
			t.Errorf("rejected override changed active time: Now() = %#v, want %#v", got, instant)
		}
	})
}
