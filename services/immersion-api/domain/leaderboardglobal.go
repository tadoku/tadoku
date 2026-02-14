package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type LeaderboardGlobalRepository interface {
	FetchGlobalLeaderboard(context.Context, *LeaderboardGlobalRequest) (*Leaderboard, error)
	FindUserDisplayNames(context.Context, []uuid.UUID) (map[uuid.UUID]string, error)
	FetchAllGlobalLeaderboardScores(context.Context) ([]LeaderboardScore, error)
}

type LeaderboardGlobalStore interface {
	FetchGlobalLeaderboardPage(ctx context.Context, page, pageSize int) ([]LeaderboardScore, int, bool, error)
	RebuildGlobalLeaderboard(ctx context.Context, scores []LeaderboardScore) error
}

type LeaderboardGlobalRequest struct {
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

type LeaderboardGlobal struct {
	repo  LeaderboardGlobalRepository
	store LeaderboardGlobalStore
}

func NewLeaderboardGlobal(repo LeaderboardGlobalRepository, store LeaderboardGlobalStore) *LeaderboardGlobal {
	return &LeaderboardGlobal{repo: repo, store: store}
}

func (s *LeaderboardGlobal) Execute(ctx context.Context, req *LeaderboardGlobalRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	if req.LanguageCode != nil || req.ActivityID != nil {
		return s.repo.FetchGlobalLeaderboard(ctx, req)
	}

	scores, totalCount, exists, err := s.store.FetchGlobalLeaderboardPage(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch global leaderboard from store: %w", err)
	}

	if !exists {
		allScores, err := s.repo.FetchAllGlobalLeaderboardScores(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all global scores for rebuild: %w", err)
		}
		if err := s.store.RebuildGlobalLeaderboard(ctx, allScores); err != nil {
			return nil, fmt.Errorf("failed to rebuild global leaderboard: %w", err)
		}
		return s.repo.FetchGlobalLeaderboard(ctx, req)
	}

	userIDs := make([]uuid.UUID, len(scores))
	for i, sc := range scores {
		userIDs[i] = sc.UserID
	}

	displayNames, err := s.repo.FindUserDisplayNames(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch display names: %w", err)
	}

	entries := buildLeaderboardEntries(scores, displayNames, req.Page*req.PageSize)

	nextPageToken := ""
	if (req.Page*req.PageSize)+req.PageSize < totalCount {
		nextPageToken = fmt.Sprint(req.Page + 1)
	}

	return &Leaderboard{
		Entries:       entries,
		TotalSize:     totalCount,
		NextPageToken: nextPageToken,
	}, nil
}
