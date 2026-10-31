package leaderboard

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Service struct {
	repository *Repository
	cache      *Cache
}

func NewService(repository *Repository, cache *Cache) *Service {
	return &Service{
		repository: repository,
		cache:      cache,
	}
}

func (s *Service) InvalidateContest(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errx.NewInvalidInputError("contest ID is required")
	}
	key, err := s.cache.cacheKey(ctx, contestPrefix+id.String())
	if err != nil {
		return err
	}

	return s.cache.invalidate(ctx, key)
}

func (s *Service) InvalidateOfficial(ctx context.Context, year int16) error {
	if year < 1 || year > 9999 {
		return errx.NewInvalidInputError("year must be between 1 and 9999")
	}
	yearlyKey, err := s.cache.cacheKey(ctx, yearlyPrefix+strconv.Itoa(int(year)))
	if err != nil {
		return err
	}
	globalCacheKey, err := s.cache.cacheKey(ctx, globalKey)
	if err != nil {
		return err
	}

	if err := s.cache.invalidate(ctx, yearlyKey); err != nil {
		return err
	}
	return s.cache.invalidate(ctx, globalCacheKey)
}

func (s *Service) FetchContest(ctx context.Context, request ContestRequest) (*Leaderboard, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	request.Request = normalize(request.Request)
	if filtered(request.Request) {
		return s.fetchContestFromPostgres(ctx, request)
	}

	key, err := s.cache.cacheKey(ctx, contestPrefix+request.ContestID.String())
	if err != nil {
		return nil, err
	}
	board, exists, err := s.cache.fetchPage(ctx, key, int64(request.offset()), request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "contest leaderboard cache unavailable; falling back to Postgres", "error", err)
		return s.fetchContestFromPostgres(ctx, request)
	}
	if !exists {
		generation, err := s.cache.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "contest leaderboard cache unavailable; falling back to Postgres", "error", err)
			return s.fetchContestFromPostgres(ctx, request)
		}
		found, err := s.repository.contestExists(ctx, request.ContestID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, errx.NewNotFoundError("contest not found")
		}
		scores, err := s.repository.allContestScores(ctx, request.ContestID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all contest scores for rebuild: %w", err)
		}
		if _, err := s.cache.rebuild(ctx, key, scores, generation); err != nil {
			slog.WarnContext(ctx, "contest leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return s.fetchContestFromPostgres(ctx, request)
	}
	return result(board, request.Request), nil
}

func (s *Service) FetchYearly(ctx context.Context, request YearlyRequest) (*Leaderboard, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	request.Request = normalize(request.Request)
	if filtered(request.Request) {
		return s.fetchYearlyFromPostgres(ctx, request)
	}

	key, err := s.cache.cacheKey(ctx, yearlyPrefix+strconv.Itoa(int(request.Year)))
	if err != nil {
		return nil, err
	}
	board, exists, err := s.cache.fetchPage(ctx, key, int64(request.offset()), request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "yearly leaderboard cache unavailable; falling back to Postgres", "error", err)
		return s.fetchYearlyFromPostgres(ctx, request)
	}
	if !exists {
		generation, err := s.cache.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "yearly leaderboard cache unavailable; falling back to Postgres", "error", err)
			return s.fetchYearlyFromPostgres(ctx, request)
		}
		scores, err := s.repository.allYearlyScores(ctx, int(request.Year))
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all yearly scores for rebuild: %w", err)
		}
		if _, err := s.cache.rebuild(ctx, key, scores, generation); err != nil {
			slog.WarnContext(ctx, "yearly leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return s.fetchYearlyFromPostgres(ctx, request)
	}
	return result(board, request.Request), nil
}

func (s *Service) FetchGlobal(ctx context.Context, request Request) (*Leaderboard, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	request = normalize(request)
	if filtered(request) {
		return s.fetchGlobalFromPostgres(ctx, request)
	}

	key, err := s.cache.cacheKey(ctx, globalKey)
	if err != nil {
		return nil, err
	}
	board, exists, err := s.cache.fetchPage(ctx, key, int64(request.offset()), request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "global leaderboard cache unavailable; falling back to Postgres", "error", err)
		return s.fetchGlobalFromPostgres(ctx, request)
	}
	if !exists {
		generation, err := s.cache.generation(ctx, key)
		if err != nil {
			slog.WarnContext(ctx, "global leaderboard cache unavailable; falling back to Postgres", "error", err)
			return s.fetchGlobalFromPostgres(ctx, request)
		}
		scores, err := s.repository.allGlobalScores(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all global scores for rebuild: %w", err)
		}
		if _, err := s.cache.rebuild(ctx, key, scores, generation); err != nil {
			slog.WarnContext(ctx, "global leaderboard cache rebuild failed; falling back to Postgres", "error", err)
		}
		return s.fetchGlobalFromPostgres(ctx, request)
	}
	return result(board, request), nil
}

func (s *Service) fetchContestFromPostgres(ctx context.Context, request ContestRequest) (*Leaderboard, error) {
	exists, err := s.repository.contestExists(ctx, request.ContestID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errx.NewNotFoundError("contest not found")
	}

	board, err := s.repository.contest(ctx, request)
	if err != nil {
		return nil, err
	}
	return result(board, request.Request), nil
}

func (s *Service) fetchYearlyFromPostgres(ctx context.Context, request YearlyRequest) (*Leaderboard, error) {
	board, err := s.repository.yearly(ctx, request)
	if err != nil {
		return nil, err
	}
	return result(board, request.Request), nil
}

func (s *Service) fetchGlobalFromPostgres(ctx context.Context, request Request) (*Leaderboard, error) {
	board, err := s.repository.global(ctx, request)
	if err != nil {
		return nil, err
	}
	return result(board, request), nil
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

func buildEntries(board *page) []Entry {
	entries := make([]Entry, len(board.scores))
	if len(entries) == 0 {
		return entries
	}

	rank := board.aboveFirst + 1
	for i, item := range board.scores {
		if i > 0 && item.value < board.scores[i-1].value {
			rank = board.offset + i + 1
		}
		entries[i] = Entry{
			Rank:   rank,
			UserID: item.userID,
			Score:  float32(item.value),
		}
	}

	for i := 1; i < len(entries); i++ {
		if entries[i].Rank == entries[i-1].Rank {
			entries[i].IsTie = true
			entries[i-1].IsTie = true
		}
	}
	if board.aboveFirst < board.offset {
		entries[0].IsTie = true
	}
	if board.throughLast > board.offset+len(entries) {
		entries[len(entries)-1].IsTie = true
	}
	return entries
}
