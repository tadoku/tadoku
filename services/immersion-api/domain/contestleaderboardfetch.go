package domain

import (
	"context"

	"github.com/google/uuid"
)

type ContestLeaderboardFetchRepository interface {
	FetchContestLeaderboard(context.Context, *ContestLeaderboardFetchRequest) (*Leaderboard, error)
}

type ContestLeaderboardFetchRequest struct {
	ContestID    uuid.UUID
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

type ContestLeaderboardFetch struct {
	repo ContestLeaderboardFetchRepository
}

func NewContestLeaderboardFetch(repo ContestLeaderboardFetchRepository) *ContestLeaderboardFetch {
	return &ContestLeaderboardFetch{repo: repo}
}

func (s *ContestLeaderboardFetch) Execute(ctx context.Context, req *ContestLeaderboardFetchRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.repo.FetchContestLeaderboard(ctx, req)
}
