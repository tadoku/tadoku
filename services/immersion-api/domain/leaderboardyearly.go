package domain

import (
	"context"
)

type LeaderboardYearlyRepository interface {
	FetchYearlyLeaderboard(context.Context, *LeaderboardYearlyRequest) (*Leaderboard, error)
}

type LeaderboardYearlyRequest struct {
	Year         int32
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

type LeaderboardYearly struct {
	repo LeaderboardYearlyRepository
}

func NewLeaderboardYearly(repo LeaderboardYearlyRepository) *LeaderboardYearly {
	return &LeaderboardYearly{repo: repo}
}

func (s *LeaderboardYearly) Execute(ctx context.Context, req *LeaderboardYearlyRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.repo.FetchYearlyLeaderboard(ctx, req)
}
