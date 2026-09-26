package leaderboard

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/leaderboardoutbox"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	valkeygo "github.com/valkey-io/valkey-go"
)

type Service struct {
	repository *Repository
	store      *Store
	cacheReady atomic.Bool
}

func NewService(repository *Repository, client valkeygo.Client, operationTimeout time.Duration, cachePrefix string) *Service {
	service := &Service{
		repository: repository,
		store:      NewStore(client, operationTimeout, cachePrefix),
	}
	service.cacheReady.Store(true)
	return service
}

func (s *Service) ReconcileCache(ctx context.Context) (int, error) { return s.store.reconcile(ctx) }

func (s *Service) InvalidateContest(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errx.NewInvalidInputError("contest ID is required")
	}
	return s.store.invalidate(ctx, s.store.cacheKey(contestPrefix+id.String()))
}

func (s *Service) InvalidateOfficial(ctx context.Context, year int16) error {
	if year < 1 {
		return errx.NewInvalidInputError("year must be positive")
	}
	if err := s.store.invalidate(ctx, s.store.cacheKey(yearlyPrefix+strconv.Itoa(int(year)))); err != nil {
		return err
	}
	return s.store.invalidate(ctx, s.store.cacheKey(globalKey))
}

func (s *Service) FetchContest(ctx context.Context, request ContestRequest) (*Result, error) {
	request.Request = normalize(request.Request)
	if err := validateActivity(request.ActivityID); err != nil {
		return nil, err
	}
	if filtered(request.Request) {
		return s.fetchContestFromPostgres(ctx, request)
	}
	if !s.cacheReady.Load() {
		return s.fetchContestFromPostgres(ctx, request)
	}

	key := s.store.cacheKey(contestPrefix + request.ContestID.String())
	result, exists, err := s.store.fetchPage(ctx, key, request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "contest leaderboard cache unavailable; falling back to Postgres", "error", err)
		return s.fetchContestFromPostgres(ctx, request)
	}
	if !exists {
		generation, err := s.store.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "contest leaderboard cache unavailable; falling back to Postgres", "error", err)
			return s.fetchContestFromPostgres(ctx, request)
		}
		scores, err := s.repository.allContestScores(ctx, request.ContestID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all contest scores for rebuild: %w", err)
		}
		if _, err := s.store.rebuild(ctx, key, scores, generation); err != nil {
			slog.WarnContext(ctx, "contest leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return s.fetchContestFromPostgres(ctx, request)
	}
	return cachedResult(result, request.Page, request.PageSize), nil
}

func (s *Service) FetchYearly(ctx context.Context, request YearlyRequest) (*Result, error) {
	request.Request = normalize(request.Request)
	if err := validateActivity(request.ActivityID); err != nil {
		return nil, err
	}
	if filtered(request.Request) {
		return postgresResult(s.repository.yearly(ctx, request))
	}
	if !s.cacheReady.Load() {
		return postgresResult(s.repository.yearly(ctx, request))
	}

	key := s.store.cacheKey(yearlyPrefix + strconv.Itoa(int(request.Year)))
	result, exists, err := s.store.fetchPage(ctx, key, request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "yearly leaderboard cache unavailable; falling back to Postgres", "error", err)
		return postgresResult(s.repository.yearly(ctx, request))
	}
	if !exists {
		generation, err := s.store.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "yearly leaderboard cache unavailable; falling back to Postgres", "error", err)
			return postgresResult(s.repository.yearly(ctx, request))
		}
		scores, err := s.repository.allYearlyScores(ctx, int(request.Year))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all yearly scores for rebuild: %w", err)
		}
		if _, err := s.store.rebuild(ctx, key, scores, generation); err != nil {
			slog.WarnContext(ctx, "yearly leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return postgresResult(s.repository.yearly(ctx, request))
	}
	return cachedResult(result, request.Page, request.PageSize), nil
}

func (s *Service) FetchGlobal(ctx context.Context, request Request) (*Result, error) {
	request = normalize(request)
	if err := validateActivity(request.ActivityID); err != nil {
		return nil, err
	}
	if filtered(request) {
		return postgresResult(s.repository.global(ctx, request))
	}
	if !s.cacheReady.Load() {
		return postgresResult(s.repository.global(ctx, request))
	}

	key := s.store.cacheKey(globalKey)
	result, exists, err := s.store.fetchPage(ctx, key, request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "global leaderboard cache unavailable; falling back to Postgres", "error", err)
		return postgresResult(s.repository.global(ctx, request))
	}
	if !exists {
		generation, err := s.store.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "global leaderboard cache unavailable; falling back to Postgres", "error", err)
			return postgresResult(s.repository.global(ctx, request))
		}
		scores, err := s.repository.allGlobalScores(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all global scores for rebuild: %w", err)
		}
		if _, err := s.store.rebuild(ctx, key, scores, generation); err != nil {
			slog.WarnContext(ctx, "global leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return postgresResult(s.repository.global(ctx, request))
	}
	return cachedResult(result, request.Page, request.PageSize), nil
}

func (s *Service) fetchContestFromPostgres(ctx context.Context, request ContestRequest) (*Result, error) {
	exists, err := s.repository.contestExists(ctx, request.ContestID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errx.NewNotFoundError("contest not found")
	}
	return postgresResult(s.repository.contest(ctx, request))
}

func normalize(request Request) Request {
	if request.PageSize == 0 {
		request.PageSize = 25
	}
	if request.PageSize > 100 {
		request.PageSize = 100
	}
	return request
}

func validateActivity(activityID *int32) error {
	if activityID == nil {
		return nil
	}
	valid := false
	for _, activity := range activities.All() {
		if activity.ID == *activityID {
			valid = true
			break
		}
	}
	if !valid {
		return errx.NewInvalidInputError(fmt.Sprintf("activity %d is not valid", *activityID))
	}
	return nil
}

func filtered(request Request) bool { return request.LanguageCode != nil || request.ActivityID != nil }

func postgresResult(value *Leaderboard, err error) (*Result, error) {
	if err != nil {
		return nil, err
	}
	return &Result{Leaderboard: value}, nil
}

func cachedResult(cached *page, currentPage, pageSize int) *Result {
	return &Result{
		Leaderboard:         result(buildEntries(cached), cached.totalCount, currentPage, pageSize),
		HydrateDisplayNames: true,
	}
}

func buildEntries(cached *page) []Entry {
	if len(cached.scores) == 0 {
		return []Entry{}
	}
	entries := make([]Entry, len(cached.scores))
	rank := cached.startRank
	for i, item := range cached.scores {
		if i > 0 && item.value < cached.scores[i-1].value {
			rank = cached.startRank + i
		}
		entries[i] = Entry{
			Rank:   rank,
			UserID: item.userID,
			Score:  float32(item.value),
		}
	}
	for i := range entries {
		if i > 0 && entries[i].Rank == entries[i-1].Rank {
			entries[i].IsTie = true
			entries[i-1].IsTie = true
		}
	}
	if cached.hasPrevTie {
		entries[0].IsTie = true
	}
	if cached.hasNextTie {
		entries[len(entries)-1].IsTie = true
	}
	return entries
}

type Worker struct {
	service    *Service
	logger     *slog.Logger
	reconciled bool
}

func NewWorker(service *Service, logger *slog.Logger) *Worker {
	service.cacheReady.Store(false)
	return &Worker{service: service, logger: logger}
}

func (w *Worker) Run(ctx context.Context) {
	poll := time.NewTicker(500 * time.Millisecond)
	defer poll.Stop()
	cleanup := time.NewTicker(time.Hour)
	defer cleanup.Stop()
	for {
		if err := w.ProcessPending(ctx); err != nil && ctx.Err() == nil {
			w.logger.ErrorContext(ctx, "leaderboard outbox pass failed", "error", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-poll.C:
		case <-cleanup.C:
			if err := w.service.repository.cleanupOutbox(ctx, timex.Now().Add(-24*time.Hour)); err != nil && ctx.Err() == nil {
				w.logger.ErrorContext(ctx, "leaderboard outbox cleanup failed", "error", err)
			}
		}
	}
}

func (w *Worker) ProcessPending(ctx context.Context) error {
	if !w.reconciled {
		count, err := w.service.store.reconcile(ctx)
		if err != nil {
			return fmt.Errorf("reconcile leaderboard cache: %w", err)
		}
		w.reconciled = true
		w.logger.InfoContext(ctx, "leaderboard cache reconciled", "invalidated", count)
	}
	if err := w.drain(ctx); err != nil {
		w.service.cacheReady.Store(false)
		return err
	}
	if !w.service.cacheReady.Swap(true) {
		w.logger.InfoContext(ctx, "leaderboard outbox ready")
	}
	return nil
}

func (w *Worker) drain(ctx context.Context) error {
	for {
		count, err := w.ProcessBatch(ctx)
		if err != nil || count < 100 {
			return err
		}
	}
}

func (w *Worker) ProcessBatch(ctx context.Context) (int, error) {
	processed := 0
	err := postgres.RunInTransaction(ctx, w.service.repository.db, func(ctx context.Context) error {
		events, err := w.service.repository.lockOutbox(ctx)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}

		keys := make(map[string]struct{})
		ids := make([]int64, 0, len(events))
		for _, event := range events {
			ids = append(ids, event.id)
			switch leaderboardoutbox.EventType(event.eventType) {
			case leaderboardoutbox.RefreshContestScore:
				if event.contestID == nil {
					w.logger.ErrorContext(ctx, "invalid leaderboard outbox event", "event_id", event.id, "event_type", event.eventType)
					continue
				}
				keys[w.service.store.cacheKey(contestPrefix+event.contestID.String())] = struct{}{}
			case leaderboardoutbox.RefreshOfficialScores:
				if event.year == nil {
					w.logger.ErrorContext(ctx, "invalid leaderboard outbox event", "event_id", event.id, "event_type", event.eventType)
					continue
				}
				keys[w.service.store.cacheKey(yearlyPrefix+strconv.Itoa(int(*event.year)))] = struct{}{}
				keys[w.service.store.cacheKey(globalKey)] = struct{}{}
			default:
				w.logger.ErrorContext(ctx, "unknown leaderboard outbox event", "event_id", event.id, "event_type", event.eventType)
			}
		}

		for key := range keys {
			if err := w.service.store.invalidate(ctx, key); err != nil {
				return err
			}
		}

		if err := w.service.repository.markOutbox(ctx, ids, timex.Now()); err != nil {
			return err
		}
		processed = len(events)
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("process leaderboard outbox: %w", err)
	}
	if processed > 0 {
		w.logger.InfoContext(ctx, "leaderboard outbox batch processed", "processed", processed)
	}
	return processed, nil
}
