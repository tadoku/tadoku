package leaderboard

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	valkeygo "github.com/valkey-io/valkey-go"
)

type Service struct {
	repository *Repository
	store      *Store
}

func NewService(repository *Repository, client valkeygo.Client, operationTimeout time.Duration, cachePrefix string) *Service {
	return &Service{
		repository: repository,
		store:      NewStore(client, operationTimeout, cachePrefix),
	}
}

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
	if err := request.Validate(); err != nil {
		return nil, err
	}
	request.Request = normalize(request.Request)
	if filtered(request.Request) {
		return s.fetchContestFromPostgres(ctx, request)
	}

	key := s.store.cacheKey(contestPrefix + request.ContestID.String())
	result, exists, err := s.store.fetchPage(ctx, key, int64(request.offset()), request.PageSize)
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
	return cachedResult(result, request.Request), nil
}

func (s *Service) FetchYearly(ctx context.Context, request YearlyRequest) (*Result, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	request.Request = normalize(request.Request)
	if filtered(request.Request) {
		return postgresResult(s.repository.yearly(ctx, request))
	}

	key := s.store.cacheKey(yearlyPrefix + strconv.Itoa(int(request.Year)))
	result, exists, err := s.store.fetchPage(ctx, key, int64(request.offset()), request.PageSize)
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
	return cachedResult(result, request.Request), nil
}

func (s *Service) FetchGlobal(ctx context.Context, request Request) (*Result, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	request = normalize(request)
	if filtered(request) {
		return postgresResult(s.repository.global(ctx, request))
	}

	key := s.store.cacheKey(globalKey)
	result, exists, err := s.store.fetchPage(ctx, key, int64(request.offset()), request.PageSize)
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
	return cachedResult(result, request), nil
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

func cachedResult(cached *page, request Request) *Result {
	return &Result{
		Leaderboard:         result(buildEntries(cached), cached.totalCount, request),
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
