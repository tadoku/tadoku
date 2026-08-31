package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ContestLeaderboardFetchRepository interface {
	FetchContestLeaderboard(context.Context, *ContestLeaderboardFetchRequest) (*Leaderboard, error)
	FindUserDisplayNames(context.Context, []uuid.UUID) (map[uuid.UUID]string, error)
	FetchAllContestLeaderboardScores(context.Context, uuid.UUID) ([]LeaderboardScore, error)
}

type ContestLeaderboardFetchStore interface {
	FetchContestLeaderboardPage(ctx context.Context, contestID uuid.UUID, page, pageSize int) (*LeaderboardPage, bool, error)
	RebuildContestLeaderboard(ctx context.Context, contestID uuid.UUID, scores []LeaderboardScore) error
}

type ContestLeaderboardFetchRequest struct {
	ContestID    uuid.UUID
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

type ContestLeaderboardFetch struct {
	repo          ContestLeaderboardFetchRepository
	store         ContestLeaderboardFetchStore
	cacheObserver LeaderboardCacheObserver
}

func NewContestLeaderboardFetch(repo ContestLeaderboardFetchRepository, store ContestLeaderboardFetchStore) *ContestLeaderboardFetch {
	return NewContestLeaderboardFetchWithCacheObserver(repo, store, nil)
}

func NewContestLeaderboardFetchWithCacheObserver(
	repo ContestLeaderboardFetchRepository,
	store ContestLeaderboardFetchStore,
	cacheObserver LeaderboardCacheObserver,
) *ContestLeaderboardFetch {
	return &ContestLeaderboardFetch{repo: repo, store: store, cacheObserver: cacheObserver}
}

func (s *ContestLeaderboardFetch) Execute(ctx context.Context, req *ContestLeaderboardFetchRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	if req.ActivityID != nil && !IsValidActivityID(*req.ActivityID) {
		return nil, fmt.Errorf("activity %d is not valid: %w", *req.ActivityID, ErrRequestInvalid)
	}

	if req.LanguageCode != nil || req.ActivityID != nil {
		return s.repo.FetchContestLeaderboard(ctx, req)
	}

	lbPage, exists, err := s.store.FetchContestLeaderboardPage(ctx, req.ContestID, req.Page, req.PageSize)
	if err != nil {
		s.observeFallback(ctx, LeaderboardCacheOperationFetch, err)
		return s.repo.FetchContestLeaderboard(ctx, req)
	}

	if !exists {
		allScores, err := s.repo.FetchAllContestLeaderboardScores(ctx, req.ContestID)
		if err != nil {
			s.observeFallback(ctx, LeaderboardCacheOperationRebuild, err)
			return s.repo.FetchContestLeaderboard(ctx, req)
		}
		if err := s.store.RebuildContestLeaderboard(ctx, req.ContestID, allScores); err != nil {
			s.observeFallback(ctx, LeaderboardCacheOperationRebuild, err)
			return s.repo.FetchContestLeaderboard(ctx, req)
		}
		s.observeFallback(ctx, LeaderboardCacheOperationFetch, nil)
		return s.repo.FetchContestLeaderboard(ctx, req)
	}

	userIDs := make([]uuid.UUID, len(lbPage.Scores))
	for i, sc := range lbPage.Scores {
		userIDs[i] = sc.UserID
	}

	displayNames, err := s.repo.FindUserDisplayNames(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch display names: %w", err)
	}

	entries := buildLeaderboardEntries(*lbPage, displayNames)

	nextPageToken := ""
	if (req.Page*req.PageSize)+req.PageSize < lbPage.TotalCount {
		nextPageToken = fmt.Sprint(req.Page + 1)
	}

	return &Leaderboard{
		Entries:       entries,
		TotalSize:     lbPage.TotalCount,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *ContestLeaderboardFetch) observeFallback(ctx context.Context, operation LeaderboardCacheOperation, err error) {
	observeLeaderboardCache(ctx, s.cacheObserver, LeaderboardCacheObservation{
		Kind:      LeaderboardCacheKindContest,
		Operation: operation,
		Outcome:   LeaderboardCacheOutcomeFallback,
		Err:       err,
	})
}
