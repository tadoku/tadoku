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

func (s *Service) FetchContest(ctx context.Context, request ContestRequest) (*Leaderboard, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}

	return s.fetch(ctx, request.Request, contestBoard(request.ContestID), s.loadContest(request))
}

// FetchContestEntry returns ErrEntryNotFound when the contest exists but the user is not on its board.
func (s *Service) FetchContestEntry(ctx context.Context, request ContestEntryRequest) (*Entry, error) {
	board, err := s.readBoard(
		ctx,
		contestBoard(request.ContestID),
		s.loadContest(ContestRequest{ContestID: request.ContestID}),
		func(scores []score) *page { return memberOf(scores, request.UserID) },
		func(ctx context.Context, key string, version int64) (*page, bool, error) {
			return s.cache.fetchMember(ctx, key, version, request.UserID)
		},
	)
	if err != nil {
		return nil, err
	}

	entries := buildEntries(board)
	if len(entries) == 0 {
		return nil, ErrEntryNotFound
	}
	return &entries[0], nil
}

func (s *Service) loadContest(request ContestRequest) func(context.Context) ([]score, error) {
	return func(ctx context.Context) ([]score, error) {
		found, err := s.repository.contestExists(ctx, request.ContestID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, errx.NewNotFoundError("contest not found")
		}
		return s.repository.contestScores(ctx, request)
	}
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

// fetch pages the cached board for an unfiltered request, and pages the scores load returns otherwise.
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

	board, err := s.readBoard(
		ctx,
		b,
		load,
		func(scores []score) *page { return pageOf(scores, request) },
		func(ctx context.Context, key string, version int64) (*page, bool, error) {
			return s.cache.fetchPage(ctx, key, version, int64(request.offset()), request.PageSize)
		},
	)
	if err != nil {
		return nil, err
	}
	return result(board, request), nil
}

// readBoard serves a page of the cached board built from the board's current version, rebuilding it on a miss, so
// every view of a board agrees. load must return every score on the board. The version is read before load, so a
// board is never published under a version newer than its scores.
func (s *Service) readBoard(
	ctx context.Context,
	b board,
	load func(context.Context) ([]score, error),
	cut func([]score) *page,
	cached func(ctx context.Context, key string, version int64) (*page, bool, error),
) (*page, error) {
	key, err := s.cache.boardKey(ctx, b)
	if err != nil {
		return nil, err
	}
	version, err := s.repository.boardVersion(ctx, b)
	if err != nil {
		return nil, err
	}

	hit, exists, err := cached(ctx, key, version)
	if err != nil {
		slog.WarnContext(ctx, "leaderboard cache unavailable; falling back to Postgres", "board", b, "error", err)
		scores, err := load(ctx)
		if err != nil {
			return nil, err
		}
		return cut(scores), nil
	}
	if exists {
		return hit, nil
	}

	scores, err := load(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.cache.rebuild(ctx, key, scores, version); err != nil {
		slog.WarnContext(ctx, "leaderboard cache rebuild failed; falling back to Postgres", "board", b, "error", err)
	}
	return cut(scores), nil
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
	sortBoard(scores)
	return cutPage(scores, int(request.offset()), request.PageSize)
}

// memberOf sorts scores into the cache's order and cuts the page holding only userID, or an empty page when userID has
// no score.
func memberOf(scores []score, userID uuid.UUID) *page {
	sortBoard(scores)

	index := slices.IndexFunc(scores, func(item score) bool { return item.userID == userID })
	if index < 0 {
		return &page{totalCount: len(scores)}
	}
	return cutPage(scores, index, 1)
}

func sortBoard(scores []score) {
	slices.SortFunc(scores, func(a, b score) int {
		if order := cmp.Compare(b.value, a.value); order != 0 {
			return order
		}
		return bytes.Compare(b.userID[:], a.userID[:])
	})
}

func cutPage(scores []score, offset, size int) *page {
	start := min(offset, len(scores))
	board := &page{
		scores:     scores[start:min(start+size, len(scores))],
		offset:     offset,
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
