package jobqueue

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestServiceRejectsInvalidInputsBeforePersistence(t *testing.T) {
	queue := NewService(nil)
	ctx := context.Background()
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	valid := jobs.InvalidateOfficialLeaderboardV1{Year: 2026}
	var typedNil *jobs.InvalidateOfficialLeaderboardV1
	cases := []struct {
		name    string
		run     func() error
		message string
	}{
		{"nil_job", func() error { return queue.Enqueue(ctx, valid, nil) }, "job must not be nil"},
		{"typed_nil_job", func() error { return queue.Enqueue(ctx, valid, typedNil) }, "job must not be nil"},
		{"invalid_year", func() error { return queue.Enqueue(ctx, valid, jobs.InvalidateOfficialLeaderboardV1{}) }, "job 1: year must be between 1 and 9999"},
		{"invalid_contest", func() error { return queue.Enqueue(ctx, valid, jobs.InvalidateContestLeaderboardV1{}) }, "job 1: contest_id must be a nonzero UUID"},
		{"claim_type", func() error { _, err := queue.Claim(ctx, "", 1, time.Second, 1); return err }, "job type must not be empty"},
		{"claim_limit_zero", func() error { _, err := queue.Claim(ctx, valid.Type(), 0, time.Second, 1); return err }, "claim limit must be between 1 and 100"},
		{"claim_limit_high", func() error { _, err := queue.Claim(ctx, valid.Type(), 101, time.Second, 1); return err }, "claim limit must be between 1 and 100"},
		{"claim_lease", func() error { _, err := queue.Claim(ctx, valid.Type(), 1, time.Nanosecond, 1); return err }, "job lease must be at least one microsecond"},
		{"claim_attempts", func() error { _, err := queue.Claim(ctx, valid.Type(), 1, time.Second, 0); return err }, "max attempts must be between 1 and 2147483647"},
		{"claim_attempts_overflow", func() error { _, err := queue.Claim(ctx, valid.Type(), 1, time.Second, math.MaxInt32+1); return err }, "max attempts must be between 1 and 2147483647"},
		{"renew_lease", func() error { _, _, err := queue.Renew(ctx, ClaimedJob{}, 0); return err }, "job lease must be at least one microsecond"},
		{"retry_past", func() error {
			_, err := queue.Retry(ctx, ClaimedJob{}, now.Add(-time.Nanosecond), "temporary", 1)
			return err
		}, "next attempt must not be before the current time"},
		{"retry_attempts", func() error { _, err := queue.Retry(ctx, ClaimedJob{}, now, "temporary", 0); return err }, "max attempts must be between 1 and 2147483647"},
		{"retry_code", func() error { _, err := queue.Retry(ctx, ClaimedJob{}, now, "secret=value", 1); return err }, "job error code must contain only lowercase letters, digits, or underscores"},
		{"fail_code_empty", func() error { _, err := queue.Fail(ctx, ClaimedJob{}, ""); return err }, "job error code must be 1 to 100 characters"},
		{"fail_code_long", func() error { _, err := queue.Fail(ctx, ClaimedJob{}, strings.Repeat("a", 101)); return err }, "job error code must be 1 to 100 characters"},
		{"replay_id", func() error { _, err := queue.Replay(ctx, 0, "operator", "reason", nil); return err }, "failed job ID must be positive"},
		{"replay_actor_empty", func() error { _, err := queue.Replay(ctx, 1, " ", "reason", nil); return err }, "replay actor must be 1 to 200 characters after trimming whitespace"},
		{"replay_actor_long", func() error { _, err := queue.Replay(ctx, 1, strings.Repeat("a", 201), "reason", nil); return err }, "replay actor must be 1 to 200 characters after trimming whitespace"},
		{"replay_reason_empty", func() error { _, err := queue.Replay(ctx, 1, "operator", " ", nil); return err }, "replay reason must be 1 to 500 characters after trimming whitespace"},
		{"replay_reason_long", func() error { _, err := queue.Replay(ctx, 1, "operator", strings.Repeat("a", 501), nil); return err }, "replay reason must be 1 to 500 characters after trimming whitespace"},
		{"cleanup_limit_zero", func() error { _, err := queue.CleanupCompleted(ctx, 0); return err }, "cleanup limit must be between 1 and 1000"},
		{"cleanup_limit_high", func() error { _, err := queue.CleanupCompleted(ctx, 1001); return err }, "cleanup limit must be between 1 and 1000"},
		{"outstanding_type", func() error { _, err := queue.Outstanding(ctx, ""); return err }, "job type must not be empty"},
		{"stats_type", func() error { _, err := queue.Stats(ctx, ""); return err }, "job type must not be empty"},
		{"unsupported_type", func() error { _, err := queue.UnsupportedStats(ctx, []jobs.Type{valid.Type(), ""}); return err }, "registered job type must not be empty"},
	}
	timex.TheWorld(now, func() {
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.run()
				if errx.KindOf(err) != errx.InvalidInput || err.Error() != tc.message {
					t.Errorf("error = %v, want invalid input %q", err, tc.message)
				}
			})
		}
	})
}

func TestEnqueueEmptyBatch(t *testing.T) {
	if err := NewService(nil).Enqueue(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupRejectsZeroBusinessTime(t *testing.T) {
	timex.TheWorld(time.Time{}, func() {
		_, err := NewService(nil).CleanupCompleted(t.Context(), 1)
		if errx.KindOf(err) != errx.InvalidInput || err.Error() != "cleanup time must not be zero" {
			t.Errorf("cleanup error = %v", err)
		}
	})
}
