package leaderboard

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"slices"

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

// InvalidateContest and InvalidateOfficial delete cached boards for the previous release, whose reads do not check
// leaderboard_versions.
func (s *Service) InvalidateContest(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errx.NewInvalidInputError("contest ID is required")
	}
	key, err := s.cache.boardKey(ctx, contestBoard(id))
	if err != nil {
		return err
	}

	return s.cache.invalidate(ctx, key)
}

func (s *Service) InvalidateOfficial(ctx context.Context, year int16) error {
	if year < 1 || year > 9999 {
		return errx.NewInvalidInputError("year must be between 1 and 9999")
	}
	yearlyKey, err := s.cache.boardKey(ctx, yearlyBoard(year))
	if err != nil {
		return err
	}
	globalCacheKey, err := s.cache.boardKey(ctx, globalBoard)
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

	load := func(ctx context.Context) ([]score, error) {
		found, err := s.repository.contestExists(ctx, request.ContestID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, errx.NewNotFoundError("contest not found")
		}
		return s.repository.contestScores(ctx, request)
	}
	return s.fetch(ctx, request.Request, contestBoard(request.ContestID), load)
}

func (s *Service) FetchYearly(ctx context.Context, request YearlyRequest) (*Leaderboard, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	year := int16(request.Year)
	load := func(ctx context.Context) ([]score, error) {
		return s.repository.officialScores(ctx, &year, request.Request)
	}
	return s.fetch(ctx, request.Request, yearlyBoard(year), load)
}

func (s *Service) FetchGlobal(ctx context.Context, request Request) (*Leaderboard, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	load := func(ctx context.Context) ([]score, error) {
		return s.repository.officialScores(ctx, nil, request)
	}
	return s.fetch(ctx, request, globalBoard, load)
}

// fetch serves an unfiltered request from the cached board built from the board's current version, rebuilding it on
// a miss, and pages the board that load returns otherwise. load must return every score on the board that matches
// the request's filters. The version is read before load, so a board is never published under a version newer than
// its scores.
func (s *Service) fetch(
	ctx context.Context,
	request Request,
	b board,
	load func(context.Context) ([]score, error),
) (*Leaderboard, error) {
	request = normalize(request)
	if filtered(request) {
		return fetchFromPostgres(ctx, request, load)
	}

	key, err := s.cache.boardKey(ctx, b)
	if err != nil {
		return nil, err
	}
	version, err := s.repository.boardVersion(ctx, b)
	if err != nil {
		return nil, err
	}
	cached, exists, err := s.cache.fetchPage(ctx, key, version, int64(request.offset()), request.PageSize)
	if err != nil {
		slog.WarnContext(ctx, "leaderboard cache unavailable; falling back to Postgres", "board", b, "error", err)
		return fetchFromPostgres(ctx, request, load)
	}
	if exists {
		return result(cached, request), nil
	}

	scores, err := load(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.cache.rebuild(ctx, key, scores, version); err != nil {
		slog.WarnContext(ctx, "leaderboard cache rebuild failed; falling back to Postgres", "board", b, "error", err)
	}
	return result(pageOf(scores, request), request), nil
}

func fetchFromPostgres(
	ctx context.Context,
	request Request,
	load func(context.Context) ([]score, error),
) (*Leaderboard, error) {
	scores, err := load(ctx)
	if err != nil {
		return nil, err
	}
	return result(pageOf(scores, request), request), nil
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

// pageOf sorts scores into the cache's order and cuts the request's page from them.
func pageOf(scores []score, request Request) *page {
	slices.SortFunc(scores, func(a, b score) int {
		if order := cmp.Compare(b.value, a.value); order != 0 {
			return order
		}
		return bytes.Compare(b.userID[:], a.userID[:])
	})

	start := min(int(request.offset()), len(scores))
	board := &page{
		scores:     scores[start:min(start+request.PageSize, len(scores))],
		offset:     int(request.offset()),
		totalCount: len(scores),
	}
	if len(board.scores) == 0 {
		return board
	}

	first, last := board.scores[0].value, board.scores[len(board.scores)-1].value
	for _, item := range scores {
		if item.value > first {
			board.aboveFirst++
		}
		if item.value >= last {
			board.throughLast++
		}
	}
	return board
}

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
