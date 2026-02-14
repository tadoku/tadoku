package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type LeaderboardYearlyRepository interface {
	FetchYearlyLeaderboard(context.Context, *LeaderboardYearlyRequest) (*Leaderboard, error)
	FindUserDisplayNames(context.Context, []uuid.UUID) (map[uuid.UUID]string, error)
	FetchAllYearlyLeaderboardScores(context.Context, int) ([]LeaderboardScore, error)
}

type LeaderboardYearlyStore interface {
	FetchYearlyLeaderboardPage(ctx context.Context, year int, page, pageSize int) (*LeaderboardPage, bool, error)
	RebuildYearlyLeaderboard(ctx context.Context, year int, scores []LeaderboardScore) error
}

type LeaderboardYearlyRequest struct {
	Year         int32
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

type LeaderboardYearly struct {
	repo  LeaderboardYearlyRepository
	store LeaderboardYearlyStore
}

func NewLeaderboardYearly(repo LeaderboardYearlyRepository, store LeaderboardYearlyStore) *LeaderboardYearly {
	return &LeaderboardYearly{repo: repo, store: store}
}

func (s *LeaderboardYearly) Execute(ctx context.Context, req *LeaderboardYearlyRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	if req.LanguageCode != nil || req.ActivityID != nil {
		return s.repo.FetchYearlyLeaderboard(ctx, req)
	}

	year := int(req.Year)
	lbPage, exists, err := s.store.FetchYearlyLeaderboardPage(ctx, year, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch yearly leaderboard from store: %w", err)
	}

	if !exists {
		allScores, err := s.repo.FetchAllYearlyLeaderboardScores(ctx, year)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch all yearly scores for rebuild: %w", err)
		}
		if err := s.store.RebuildYearlyLeaderboard(ctx, year, allScores); err != nil {
			return nil, fmt.Errorf("failed to rebuild yearly leaderboard: %w", err)
		}
		return s.repo.FetchYearlyLeaderboard(ctx, req)
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
