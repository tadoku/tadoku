package query

import (
	"context"

	"github.com/google/uuid"
)

type FetchContestLeaderboardRequest struct {
	ContestID    uuid.UUID
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

func (s *ServiceImpl) FetchContestLeaderboard(ctx context.Context, req *FetchContestLeaderboardRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.r.FetchContestLeaderboard(ctx, req)
}
