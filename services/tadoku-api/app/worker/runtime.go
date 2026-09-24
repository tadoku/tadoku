package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/asyncwork"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	"github.com/tadoku/tadoku/services/tadoku-api/storage/postgres/asyncoutbox"
)

const globalLimit = 4

type policy struct {
	typeName    asyncwork.Type
	limit       int
	timeout     time.Duration
	maxRuntime  time.Duration
	lease       time.Duration
	maxAttempts int
}

var policies = [...]policy{
	{asyncwork.InvalidateContest, 2, 20 * time.Second, 30 * time.Second, 10 * time.Second, 5},
	{asyncwork.InvalidateOfficial, 2, 20 * time.Second, 30 * time.Second, 10 * time.Second, 5},
}

type Runner struct {
	repository      *asyncoutbox.Repository
	application     *Application
	leaderboard     *leaderboard.Service
	logger          *slog.Logger
	metrics         *Metrics
	shutdownTimeout time.Duration
	ready           atomic.Bool
}

func NewRunner(repository *asyncoutbox.Repository, application *Application, leaderboard *leaderboard.Service, logger *slog.Logger, metrics *Metrics, shutdownTimeout time.Duration) *Runner {
	return &Runner{repository: repository, application: application, leaderboard: leaderboard, logger: logger, metrics: metrics, shutdownTimeout: shutdownTimeout}
}

func (r *Runner) Ready() bool { return r.ready.Load() }

type result struct {
	typeName asyncwork.Type
	err      error
}

func (r *Runner) Run(ctx context.Context) {
	workCtx, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()

	if err := r.leaderboard.RevokeCacheReadiness(workCtx); err != nil {
		r.logger.Warn("revoke leaderboard readiness at startup", "error", err)
	}

	results := make(chan result, globalLimit)
	active := make(map[asyncwork.Type]int, len(policies))
	var running sync.WaitGroup
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	cleanupTicker := time.NewTicker(time.Hour)
	defer cleanupTicker.Stop()
	rotation := 0
	reconciled := false
	unsupported := int64(0)

	for {
		if ctx.Err() != nil {
			break
		}
		if !reconciled && totalActive(active) == 0 {
			count, err := r.leaderboard.ReconcileCache(ctx)
			if err != nil {
				r.logger.Error("reconcile leaderboard cache", "error", err)
			} else {
				reconciled = true
				r.logger.Info("leaderboard cache reconciled", "invalidated", count)
			}
		}

		claimHealthy := true
		for offset := range policies {
			index := (rotation + offset) % len(policies)
			spec := policies[index]
			free := min(globalLimit-totalActive(active), spec.limit-active[spec.typeName])
			if free < 1 {
				continue
			}
			tasks, err := r.repository.Claim(ctx, spec.typeName, free, spec.lease, spec.maxAttempts)
			if err != nil {
				claimHealthy = false
				r.logger.Error("claim async work", "type", spec.typeName, "error", err)
				continue
			}
			for _, task := range tasks {
				if task.Reclaimed {
					r.metrics.ExpiredLeases.WithLabelValues(string(spec.typeName)).Inc()
				}
				active[spec.typeName]++
				r.metrics.InFlight.WithLabelValues(string(spec.typeName)).Inc()
				running.Add(1)
				go func() {
					defer running.Done()
					err := r.process(workCtx, task, spec)
					results <- result{typeName: spec.typeName, err: err}
				}()
			}
		}
		rotation = (rotation + 1) % len(policies)
		r.ready.Store(claimHealthy)

		if reconciled && totalActive(active) == 0 {
			if err := r.refreshReadiness(ctx); err != nil {
				r.logger.Error("refresh leaderboard readiness", "error", err)
			}
		} else if err := r.leaderboard.RevokeCacheReadiness(workCtx); err != nil {
			r.logger.Warn("revoke leaderboard readiness while work is active", "error", err)
		}
		count, err := r.updateMetrics(ctx)
		if err != nil {
			r.logger.Warn("inspect async work backlog", "error", err)
		} else if count != unsupported {
			unsupported = count
			if count > 0 {
				r.logger.Error("unsupported async task types remain pending", "count", count)
			}
		}

		select {
		case <-ctx.Done():
		case completed := <-results:
			active[completed.typeName]--
			r.metrics.InFlight.WithLabelValues(string(completed.typeName)).Dec()
			if completed.err != nil {
				reconciled = false
				if err := r.leaderboard.RevokeCacheReadiness(workCtx); err != nil {
					r.logger.Warn("revoke leaderboard readiness after task failure", "error", err)
				}
			}
		case <-ticker.C:
		case <-cleanupTicker.C:
			r.cleanupCompleted(ctx)
		}
	}

	r.ready.Store(false)
	shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	if err := r.leaderboard.RevokeCacheReadiness(shutdownCtx); err != nil {
		r.logger.Warn("revoke leaderboard readiness at shutdown", "error", err)
	}
	stop()

	done := make(chan struct{})
	go func() { running.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(r.shutdownTimeout):
		cancelWork()
		<-done
	}
}

func (r *Runner) cleanupCompleted(ctx context.Context) {
	ctx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	for {
		count, err := r.repository.CleanupCompleted(ctx, timex.Now().Add(-24*time.Hour), 100)
		if err != nil {
			if ctx.Err() == nil {
				r.logger.Warn("cleanup completed async tasks", "error", err)
			}
			return
		}
		if count < 100 {
			return
		}
	}
}

func totalActive(active map[asyncwork.Type]int) int {
	total := 0
	for _, count := range active {
		total += count
	}
	return total
}

func (r *Runner) refreshReadiness(ctx context.Context) error {
	for _, spec := range policies {
		count, err := r.repository.Outstanding(ctx, spec.typeName)
		if err != nil {
			_ = r.leaderboard.RevokeCacheReadiness(ctx)
			return err
		}
		if count > 0 {
			return r.leaderboard.RevokeCacheReadiness(ctx)
		}
	}
	return r.leaderboard.PublishCacheReadiness(ctx)
}

func (r *Runner) updateMetrics(ctx context.Context) (int64, error) {
	known := make([]asyncwork.Type, 0, len(policies))
	for _, spec := range policies {
		known = append(known, spec.typeName)
		stats, err := r.repository.Stats(ctx, spec.typeName)
		if err != nil {
			return 0, err
		}
		r.metrics.Pending.WithLabelValues(string(spec.typeName)).Set(float64(stats.Pending))
		r.metrics.Failed.WithLabelValues(string(spec.typeName)).Set(float64(stats.Failed))
		age := float64(0)
		if stats.OldestDueAt != nil {
			age = max(0, timex.Now().Sub(*stats.OldestDueAt).Seconds())
		}
		r.metrics.OldestDueAge.WithLabelValues(string(spec.typeName)).Set(age)
	}
	stats, err := r.repository.UnsupportedStats(ctx, known)
	if err != nil {
		return 0, err
	}
	r.metrics.Unsupported.Set(float64(stats.Pending))
	age := float64(0)
	if stats.OldestDueAt != nil {
		age = max(0, timex.Now().Sub(*stats.OldestDueAt).Seconds())
	}
	r.metrics.UnsupportedOldestDueAge.Set(age)
	return stats.Pending, nil
}

func (r *Runner) process(ctx context.Context, task asyncoutbox.ClaimedTask, spec policy) error {
	margin := spec.lease / 3
	remainingLease := time.Until(task.LeaseExpiresAt)
	if remainingLease <= margin {
		return fmt.Errorf("task %d: claim lease expired before dispatch", task.ID)
	}
	if remainingLease < 2*margin {
		renewCtx, stop := context.WithTimeout(ctx, remainingLease-margin)
		expiry, held, err := r.repository.Renew(renewCtx, task, spec.lease)
		stop()
		if err != nil {
			return fmt.Errorf("renew claim before dispatch: %w", err)
		}
		if !held {
			return fmt.Errorf("task %d: claim lease lost before dispatch", task.ID)
		}
		task.LeaseExpiresAt = expiry
	}
	limit := min(spec.timeout, spec.maxRuntime)
	handlerCtx, cancelHandler := context.WithTimeout(ctx, limit)
	maximum, _ := handlerCtx.Deadline()
	renewCtx, cancelRenew := context.WithCancel(handlerCtx)
	renewed := make(chan error, 1)
	go func() { renewed <- r.renew(renewCtx, cancelHandler, task, spec, maximum) }()

	handlerErr := r.application.Dispatch(handlerCtx, task)
	if handlerErr == nil && handlerCtx.Err() != nil {
		handlerErr = handlerCtx.Err()
	}
	cancelRenew()
	renewErr := <-renewed
	cancelHandler()
	if renewErr != nil {
		r.logger.Warn("async work lease lost", "task_id", task.ID, "type", task.Type, "error", renewErr)
		return renewErr
	}

	r.metrics.Duration.WithLabelValues(string(task.Type)).Observe(max(0, (limit - time.Until(maximum)).Seconds()))
	transitionCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	var held bool
	var err error
	if handlerErr == nil {
		held, err = r.repository.Complete(transitionCtx, task)
	} else {
		code := failureCode(handlerErr)
		var permanent *PermanentError
		if errors.As(handlerErr, &permanent) {
			held, err = r.repository.Fail(transitionCtx, task, code)
		} else {
			held, err = r.repository.Retry(transitionCtx, task, retryAt(task.Attempts), code, spec.maxAttempts)
		}
		r.metrics.Attempts.WithLabelValues(string(task.Type), code).Inc()
	}
	if err != nil {
		return fmt.Errorf("transition task %d: %w", task.ID, err)
	}
	if !held {
		return fmt.Errorf("task %d: lease lost before transition", task.ID)
	}
	if handlerErr != nil {
		r.logger.Error("async work failed", "task_id", task.ID, "type", task.Type, "attempt", task.Attempts, "code", failureCode(handlerErr), "error", handlerErr)
	}
	return handlerErr
}

func (r *Runner) renew(ctx context.Context, cancelHandler context.CancelFunc, task asyncoutbox.ClaimedTask, spec policy, maximum time.Time) error {
	leaseExpiresAt := task.LeaseExpiresAt
	for {
		if ctx.Err() != nil {
			return nil
		}
		remainingLease := time.Until(leaseExpiresAt)
		margin := spec.lease / 3
		if remainingLease <= margin {
			if ctx.Err() != nil {
				return nil
			}
			cancelHandler()
			return context.DeadlineExceeded
		}
		wait := min(margin, remainingLease-margin)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
			remaining := time.Until(maximum)
			renewBudget := time.Until(leaseExpiresAt) - margin
			if remaining <= 0 || renewBudget <= 0 {
				cancelHandler()
				return context.DeadlineExceeded
			}
			renewCtx, stop := context.WithTimeout(ctx, renewBudget)
			updatedExpiry, held, err := r.repository.Renew(renewCtx, task, min(spec.lease, remaining))
			stop()
			if ctx.Err() != nil {
				return nil
			}
			if err != nil || !held {
				cancelHandler()
				if err != nil {
					return err
				}
				return errors.New("lease token no longer matches")
			}
			leaseExpiresAt = updatedExpiry
		}
	}
}

func retryAt(attempts int) time.Time {
	shift := min(max(attempts-1, 0), 6)
	backoff := min(time.Second<<shift, time.Minute)
	jitter := time.Duration(rand.Int64N(int64(backoff / 2)))
	return timex.Now().Add(backoff/2 + jitter)
}
