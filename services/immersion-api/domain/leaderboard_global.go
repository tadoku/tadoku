package domain

import (
	"context"
)

type LeaderboardGlobalRepository interface {
	FetchGlobalLeaderboard(context.Context, *LeaderboardGlobalRequest) (*Leaderboard, error)
}

type LeaderboardGlobalRequest struct {
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

type LeaderboardGlobal struct {
	repo LeaderboardGlobalRepository
}

func NewLeaderboardGlobal(repo LeaderboardGlobalRepository) *LeaderboardGlobal {
	return &LeaderboardGlobal{repo: repo}
}

func (s *LeaderboardGlobal) Execute(ctx context.Context, req *LeaderboardGlobalRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.repo.FetchGlobalLeaderboard(ctx, req)
}
