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

	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type handlerSpec struct {
	typeName    jobs.Type
	limit       int
	timeout     time.Duration
	lease       time.Duration
	maxAttempts int
}

type runner struct {
	queue           *jobqueue.Service
	handlers        *registry
	logger          *slog.Logger
	metrics         *Metrics
	shutdownTimeout time.Duration
	concurrency     int
	ready           atomic.Bool
}

func (r *runner) run(ctx context.Context) {
	workCtx, cancelWork := context.WithCancel(context.Background())
	defer cancelWork()

	capacity := 0
	for _, entry := range r.handlers.ordered {
		capacity += entry.spec.limit
	}
	results := make(chan jobs.Type, min(r.concurrency, capacity))
	active := make(map[jobs.Type]int, len(r.handlers.ordered))
	var running sync.WaitGroup
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	metricsTicker := time.NewTicker(15 * time.Second)
	defer metricsTicker.Stop()
	cleanupTicker := time.NewTicker(time.Hour)
	defer cleanupTicker.Stop()
	rotation := 0
	unsupported := int64(0)
	refreshMetrics := func() {
		count, err := r.updateMetrics(ctx)
		if err != nil {
			r.logger.Warn("inspect job backlog", "error", err)
		} else if count != unsupported {
			unsupported = count
			if count > 0 {
				r.logger.Error("unsupported job types remain outstanding", "count", count)
			}
		}
	}

	r.cleanupCompleted(ctx)
	refreshMetrics()
	for {
		if ctx.Err() != nil {
			break
		}

		claimHealthy := true
		for offset := range r.handlers.ordered {
			index := (rotation + offset) % len(r.handlers.ordered)
			spec := r.handlers.ordered[index].spec
			free := min(r.concurrency-totalActive(active), spec.limit-active[spec.typeName])
			if free < 1 {
				continue
			}
			tasks, err := r.queue.Claim(ctx, spec.typeName, free, spec.lease, spec.maxAttempts)
			if err != nil {
				claimHealthy = false
				r.logger.Error("claim jobs", "type", spec.typeName, "error", err)
				continue
			}
			for _, task := range tasks {
				if task.Reclaimed {
					r.metrics.expiredLeases.WithLabelValues(string(spec.typeName)).Inc()
				}
				active[spec.typeName]++
				r.metrics.inFlight.WithLabelValues(string(spec.typeName)).Inc()
				running.Add(1)
				go func() {
					defer running.Done()
					_ = r.process(workCtx, task, spec)
					results <- spec.typeName
				}()
			}
		}
		rotation = (rotation + 1) % len(r.handlers.ordered)
		r.ready.Store(claimHealthy)

		select {
		case <-ctx.Done():
		case completed := <-results:
			active[completed]--
			r.metrics.inFlight.WithLabelValues(string(completed)).Dec()
		case <-ticker.C:
		case <-metricsTicker.C:
			refreshMetrics()
		case <-cleanupTicker.C:
			r.cleanupCompleted(ctx)
		}
	}

	r.ready.Store(false)

	done := make(chan struct{})
	go func() { running.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(r.shutdownTimeout):
		cancelWork()
		<-done
	}
}

func (r *runner) cleanupCompleted(ctx context.Context) {
	ctx, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	for {
		count, err := r.queue.CleanupCompleted(ctx, 100)
		if err != nil {
			if ctx.Err() == nil {
				r.logger.Warn("cleanup completed jobs", "error", err)
			}
			return
		}
		if count < 100 {
			return
		}
	}
}

func totalActive(active map[jobs.Type]int) int {
	total := 0
	for _, count := range active {
		total += count
	}
	return total
}

func (r *runner) updateMetrics(ctx context.Context) (int64, error) {
	known := make([]jobs.Type, 0, len(r.handlers.ordered))
	for _, entry := range r.handlers.ordered {
		spec := entry.spec
		known = append(known, spec.typeName)
		stats, err := r.queue.Stats(ctx, spec.typeName)
		if err != nil {
			return 0, err
		}
		r.metrics.pending.WithLabelValues(string(spec.typeName)).Set(float64(stats.Pending))
		r.metrics.failed.WithLabelValues(string(spec.typeName)).Set(float64(stats.Failed))
		age := float64(0)
		if stats.OldestDueAt != nil {
			age = max(0, timex.Now().Sub(*stats.OldestDueAt).Seconds())
		}
		r.metrics.oldestDueAge.WithLabelValues(string(spec.typeName)).Set(age)
	}
	stats, err := r.queue.UnsupportedStats(ctx, known)
	if err != nil {
		return 0, err
	}
	r.metrics.unsupported.Set(float64(stats.Pending))
	r.metrics.unsupportedRunning.Set(float64(stats.Running))
	r.metrics.unsupportedFailed.Set(float64(stats.Failed))
	age := float64(0)
	if stats.OldestDueAt != nil {
		age = max(0, timex.Now().Sub(*stats.OldestDueAt).Seconds())
	}
	r.metrics.unsupportedOldestDueAge.Set(age)
	return stats.Pending + stats.Running + stats.Failed, nil
}

func (r *runner) process(ctx context.Context, task jobqueue.ClaimedJob, spec handlerSpec) error {
	margin := spec.lease / 3
	remainingLease := time.Until(task.LeaseExpiresAt)
	if remainingLease <= margin {
		return fmt.Errorf("job %d: claim lease expired before dispatch", task.ID)
	}
	if remainingLease < 2*margin {
		renewCtx, stop := context.WithTimeout(ctx, remainingLease-margin)
		expiry, held, err := r.queue.Renew(renewCtx, task, spec.lease)
		stop()
		if err != nil {
			return fmt.Errorf("renew claim before dispatch: %w", err)
		}
		if !held {
			return fmt.Errorf("job %d: claim lease lost before dispatch", task.ID)
		}
		task.LeaseExpiresAt = expiry
	}
	limit := spec.timeout
	handlerCtx, cancelHandler := context.WithTimeout(ctx, limit)
	maximum, _ := handlerCtx.Deadline()
	renewCtx, cancelRenew := context.WithCancel(handlerCtx)
	renewed := make(chan error, 1)
	go func() { renewed <- r.renew(renewCtx, cancelHandler, task, spec, maximum) }()

	handlerErr := r.handlers.dispatch(handlerCtx, task)
	if handlerErr == nil && handlerCtx.Err() != nil {
		handlerErr = handlerCtx.Err()
	}
	cancelRenew()
	renewErr := <-renewed
	cancelHandler()
	if renewErr != nil {
		r.logger.Warn("job lease lost", "job_id", task.ID, "type", task.Type, "error", renewErr)
		return renewErr
	}

	r.metrics.duration.WithLabelValues(string(task.Type)).Observe(max(0, (limit - time.Until(maximum)).Seconds()))
	transitionCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()

	var held bool
	var err error
	if handlerErr == nil {
		held, err = r.queue.Complete(transitionCtx, task)
	} else {
		code := failureCode(handlerErr)
		var permanent *permanentError
		if errors.As(handlerErr, &permanent) {
			held, err = r.queue.Fail(transitionCtx, task, code)
		} else {
			held, err = r.queue.Retry(transitionCtx, task, retryAt(task.Attempts), code, spec.maxAttempts)
		}
		r.metrics.attempts.WithLabelValues(string(task.Type), code).Inc()
	}
	if err != nil {
		return fmt.Errorf("transition job %d: %w", task.ID, err)
	}
	if !held {
		return fmt.Errorf("job %d: lease lost before transition", task.ID)
	}
	if handlerErr != nil {
		r.logger.Error("job failed", "job_id", task.ID, "type", task.Type, "attempt", task.Attempts, "code", failureCode(handlerErr), "error", handlerErr)
	}
	return handlerErr
}

func (r *runner) renew(ctx context.Context, cancelHandler context.CancelFunc, task jobqueue.ClaimedJob, spec handlerSpec, maximum time.Time) error {
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
			if remaining <= 0 {
				return nil
			}
			if renewBudget <= 0 {
				cancelHandler()
				return context.DeadlineExceeded
			}
			renewCtx, stop := context.WithTimeout(ctx, renewBudget)
			updatedExpiry, held, err := r.queue.Renew(renewCtx, task, spec.lease)
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
