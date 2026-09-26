package worker

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
)

func TestRegistryRejectsInvalidRegistrations(t *testing.T) {
	valid := Policy{Concurrency: 2, Timeout: time.Second, MaxAttempts: 5}
	handler := func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error { return nil }
	for _, tc := range []struct {
		name    string
		entries []registration
	}{
		{"empty", nil},
		{"duplicate", []registration{handle(handler, valid), handle(handler, valid)}},
		{"nil handler", []registration{handle[jobs.InvalidateOfficialLeaderboardV1](nil, valid)}},
		{"pointer payload", []registration{handle(func(context.Context, *jobs.InvalidateOfficialLeaderboardV1) error { return nil }, valid)}},
		{"interface payload", []registration{handle(func(context.Context, jobs.Job) error { return nil }, valid)}},
		{"zero concurrency", []registration{handle(handler, Policy{Timeout: time.Second, MaxAttempts: 5})}},
		{"negative concurrency", []registration{handle(handler, Policy{Concurrency: -1, Timeout: time.Second, MaxAttempts: 5})}},
		{"oversized concurrency", []registration{handle(handler, Policy{Concurrency: 101, Timeout: time.Second, MaxAttempts: 5})}},
		{"oversized attempts", []registration{handle(handler, Policy{Concurrency: 2, Timeout: time.Second, MaxAttempts: math.MaxInt32 + 1})}},
		{"zero timeout", []registration{handle(handler, Policy{Concurrency: 2, MaxAttempts: 5})}},
		{"zero attempts", []registration{handle(handler, Policy{Concurrency: 2, Timeout: time.Second})}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := newRegistry(tc.entries...); err == nil {
				t.Fatal("invalid registration accepted")
			}
		})
	}
}

func TestRegistryValidatesPayloadBeforeCallingTypedHandler(t *testing.T) {
	called := 0
	var got jobs.InvalidateOfficialLeaderboardV1
	var gotCtx context.Context
	handlers, err := newRegistry(handle(func(ctx context.Context, job jobs.InvalidateOfficialLeaderboardV1) error {
		called++
		got, gotCtx = job, ctx
		return nil
	}, Policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 1}))
	if err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{`{"year":0}`, `{"year":2025,"unexpected":true}`, `{"year":2025} {}`, `{"year":`, `null`, `[]`, `"text"`} {
		err := handlers.dispatch(t.Context(), jobqueue.ClaimedJob{Type: jobs.InvalidateOfficial, Payload: []byte(raw)})
		var permanent *PermanentError
		if !errors.As(err, &permanent) {
			t.Errorf("payload %s: got %v; want permanent failure", raw, err)
		}
	}
	if called != 0 {
		t.Fatalf("invalid payload reached handler %d times", called)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	if err := handlers.dispatch(ctx, jobqueue.ClaimedJob{Type: jobs.InvalidateOfficial, Payload: []byte(`{"year":2025}`)}); err != nil {
		t.Fatal(err)
	}
	if called != 1 || got.Year != 2025 || gotCtx != ctx {
		t.Errorf("typed invocation: calls=%d job=%+v context preserved=%t", called, got, gotCtx == ctx)
	}
	err = handlers.dispatch(ctx, jobqueue.ClaimedJob{Type: "future.task.v1"})
	var unknown *UnknownTypeError
	if !errors.As(err, &unknown) || called != 1 {
		t.Errorf("unknown job: error=%v calls=%d", err, called)
	}
}
