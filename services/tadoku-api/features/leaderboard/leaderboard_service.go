package leaderboard

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
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

const (
	contestPrefix = "leaderboard:contest:"
	yearlyPrefix  = "leaderboard:yearly:"
	globalKey     = "leaderboard:global"
)

var rebuildScript = valkeygo.NewLuaScript(`
if (redis.call('GET', KEYS[3]) or '0') ~= ARGV[1] then
  return 0
end
redis.call('DEL', KEYS[1])
if #ARGV > 1 then
  redis.call('ZADD', KEYS[1], unpack(ARGV, 2))
end
redis.call('SET', KEYS[2], 'native:' .. ARGV[1])
return 1
`)

var invalidateScript = valkeygo.NewLuaScript(`
redis.call('INCR', KEYS[3])
redis.call('DEL', KEYS[1], KEYS[2])
return 1
`)

type Service struct {
	repository       *Repository
	client           valkeygo.Client
	operationTimeout time.Duration
	cachePrefix      string
	cacheReady       atomic.Bool
}

func NewService(repository *Repository, client valkeygo.Client, operationTimeout time.Duration, cachePrefix string) *Service {
	service := &Service{
		repository:       repository,
		client:           client,
		operationTimeout: operationTimeout,
		cachePrefix:      cachePrefix,
	}
	service.cacheReady.Store(true)
	return service
}

func (s *Service) cacheKey(key string) string { return s.cachePrefix + key }

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

	key := s.cacheKey(contestPrefix + request.ContestID.String())
	result, exists, err := s.fetchPage(ctx, key, request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "contest leaderboard cache unavailable; falling back to Postgres", "error", err)
		return s.fetchContestFromPostgres(ctx, request)
	}
	if !exists {
		generation, err := s.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "contest leaderboard cache unavailable; falling back to Postgres", "error", err)
			return s.fetchContestFromPostgres(ctx, request)
		}
		scores, err := s.repository.allContestScores(ctx, request.ContestID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all contest scores for rebuild: %w", err)
		}
		if _, err := s.rebuild(ctx, key, scores, generation); err != nil {
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

	key := s.cacheKey(yearlyPrefix + strconv.Itoa(int(request.Year)))
	result, exists, err := s.fetchPage(ctx, key, request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "yearly leaderboard cache unavailable; falling back to Postgres", "error", err)
		return postgresResult(s.repository.yearly(ctx, request))
	}
	if !exists {
		generation, err := s.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "yearly leaderboard cache unavailable; falling back to Postgres", "error", err)
			return postgresResult(s.repository.yearly(ctx, request))
		}
		scores, err := s.repository.allYearlyScores(ctx, int(request.Year))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all yearly scores for rebuild: %w", err)
		}
		if _, err := s.rebuild(ctx, key, scores, generation); err != nil {
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

	key := s.cacheKey(globalKey)
	result, exists, err := s.fetchPage(ctx, key, request.Page, request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "global leaderboard cache unavailable; falling back to Postgres", "error", err)
		return postgresResult(s.repository.global(ctx, request))
	}
	if !exists {
		generation, err := s.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "global leaderboard cache unavailable; falling back to Postgres", "error", err)
			return postgresResult(s.repository.global(ctx, request))
		}
		scores, err := s.repository.allGlobalScores(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all global scores for rebuild: %w", err)
		}
		if _, err := s.rebuild(ctx, key, scores, generation); err != nil {
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

func (s *Service) fetchPage(ctx context.Context, key string, currentPage, pageSize int) (*page, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	marker, err := s.client.Do(ctx, s.client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err == valkeygo.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("check leaderboard marker %s: %w", key, err)
	}
	generation, err := s.generation(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if generation != "0" && marker != "native:"+generation {
		return nil, false, nil
	}
	total, err := s.client.Do(ctx, s.client.B().Zcard().Key(key).Build()).AsInt64()
	if err != nil {
		return nil, false, fmt.Errorf("get leaderboard cardinality %s: %w", key, err)
	}
	if total == 0 {
		current, err := s.cacheCurrent(ctx, key, marker, generation)
		if err != nil || !current {
			return nil, false, err
		}
		return &page{scores: []score{}, startRank: 1}, true, nil
	}

	start := int64(currentPage * pageSize)
	stop := start + int64(pageSize) - 1
	fetchStart := start
	if fetchStart > 0 {
		fetchStart--
	}
	values, err := s.client.Do(ctx, s.client.B().Zrange().Key(key).Min(strconv.FormatInt(fetchStart, 10)).Max(strconv.FormatInt(stop+1, 10)).Rev().Withscores().Build()).AsZScores()
	if err != nil {
		return nil, false, fmt.Errorf("fetch leaderboard page %s: %w", key, err)
	}
	hasPrevious := fetchStart < start && len(values) > 0
	previous := float64(0)
	if hasPrevious {
		previous = values[0].Score
		values = values[1:]
	}
	hasNext := int64(len(values)) > int64(pageSize)
	next := float64(0)
	if hasNext {
		next = values[len(values)-1].Score
		values = values[:len(values)-1]
	}
	scores := make([]score, len(values))
	for i, value := range values {
		id, err := uuid.Parse(value.Member)
		if err != nil {
			return nil, false, fmt.Errorf("parse leaderboard member %q: %w", value.Member, err)
		}
		scores[i] = score{userID: id, value: value.Score}
	}
	startRank := int(start) + 1
	if len(scores) > 0 {
		higher, err := s.client.Do(ctx, s.client.B().Zcount().Key(key).Min("("+strconv.FormatFloat(scores[0].value, 'f', -1, 64)).Max("+inf").Build()).AsInt64()
		if err != nil {
			return nil, false, fmt.Errorf("count higher leaderboard scores %s: %w", key, err)
		}
		startRank = int(higher) + 1
	}
	current, err := s.cacheCurrent(ctx, key, marker, generation)
	if err != nil || !current {
		return nil, false, err
	}
	return &page{
		scores:     scores,
		totalCount: int(total),
		startRank:  startRank,
		hasPrevTie: hasPrevious && len(scores) > 0 && previous == scores[0].value,
		hasNextTie: hasNext && len(scores) > 0 && next == scores[len(scores)-1].value,
	}, true, nil
}

func (s *Service) cacheCurrent(ctx context.Context, key, marker, generation string) (bool, error) {
	currentMarker, err := s.client.Do(ctx, s.client.B().Get().Key(key+":last_updated").Build()).ToString()
	if err == valkeygo.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("recheck leaderboard marker %s: %w", key, err)
	}
	if currentMarker != marker {
		return false, nil
	}
	currentGeneration, err := s.generation(ctx, key)
	if err != nil {
		return false, err
	}
	return currentGeneration == generation, nil
}

func (s *Service) generation(ctx context.Context, key string) (string, error) {
	value, err := s.client.Do(ctx, s.client.B().Get().Key(key+":generation").Build()).ToString()
	if err == valkeygo.Nil {
		return "0", nil
	}
	if err != nil {
		return "", fmt.Errorf("read leaderboard generation %s: %w", key, err)
	}
	return value, nil
}

func (s *Service) invalidate(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	if err := invalidateScript.Exec(ctx, s.client, []string{key, key + ":last_updated", key + ":generation"}, nil).Error(); err != nil {
		return fmt.Errorf("invalidate leaderboard %s: %w", key, err)
	}
	return nil
}

func (s *Service) reconcile(ctx context.Context) (int, error) {
	var cursor uint64
	count := 0
	for {
		commandCtx, cancel := context.WithTimeout(ctx, s.operationTimeout)
		page, err := s.client.Do(commandCtx, s.client.B().Scan().Cursor(cursor).Match(s.cacheKey("leaderboard:*:last_updated")).Count(100).Build()).AsScanEntry()
		cancel()
		if err != nil {
			return count, fmt.Errorf("scan leaderboard cache markers: %w", err)
		}
		for _, marker := range page.Elements {
			key := strings.TrimSuffix(marker, ":last_updated")
			if !s.leaderboardCacheKey(key) {
				continue
			}
			if err := s.invalidate(ctx, key); err != nil {
				return count, err
			}
			count++
		}
		if page.Cursor == 0 {
			return count, nil
		}
		cursor = page.Cursor
	}
}

func (s *Service) leaderboardCacheKey(key string) bool {
	if !strings.HasPrefix(key, s.cachePrefix) {
		return false
	}
	key = strings.TrimPrefix(key, s.cachePrefix)
	if key == globalKey {
		return true
	}
	if strings.HasPrefix(key, yearlyPrefix) {
		_, err := strconv.Atoi(strings.TrimPrefix(key, yearlyPrefix))
		return err == nil
	}
	if strings.HasPrefix(key, contestPrefix) {
		_, err := uuid.Parse(strings.TrimPrefix(key, contestPrefix))
		return err == nil
	}
	return false
}

func (s *Service) rebuild(ctx context.Context, key string, scores []score, generation string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, s.operationTimeout)
	defer cancel()
	args := make([]string, 0, 1+len(scores)*2)
	args = append(args, generation)
	for _, item := range scores {
		args = append(args, strconv.FormatFloat(item.value, 'f', -1, 64), item.userID.String())
	}
	published, err := rebuildScript.Exec(ctx, s.client, []string{key, key + ":last_updated", key + ":generation"}, args).ToInt64()
	if err != nil {
		return false, fmt.Errorf("rebuild leaderboard %s: %w", key, err)
	}
	return published == 1, nil
}

// Worker consumes committed leaderboard changes after the legacy worker has stopped.
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

// ProcessPending reconciles the cache once, then drains committed outbox rows.
func (w *Worker) ProcessPending(ctx context.Context) error {
	if !w.reconciled {
		count, err := w.service.reconcile(ctx)
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

// ProcessBatch claims rows through commit, then acknowledges only after Valkey succeeds.
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
				keys[w.service.cacheKey(contestPrefix+event.contestID.String())] = struct{}{}
			case leaderboardoutbox.RefreshOfficialScores:
				if event.year == nil {
					w.logger.ErrorContext(ctx, "invalid leaderboard outbox event", "event_id", event.id, "event_type", event.eventType)
					continue
				}
				keys[w.service.cacheKey(yearlyPrefix+strconv.Itoa(int(*event.year)))] = struct{}{}
				keys[w.service.cacheKey(globalKey)] = struct{}{}
			default:
				w.logger.ErrorContext(ctx, "unknown leaderboard outbox event", "event_id", event.id, "event_type", event.eventType)
			}
		}

		// Invalidate while the claimed rows stay locked so no other worker acknowledges them first.
		for key := range keys {
			if err := w.service.invalidate(ctx, key); err != nil {
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
