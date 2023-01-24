package query

import (
	"context"
)

type FetchYearlyLeaderboardRequest struct {
	Year         int
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

func (s *ServiceImpl) FetchYearlyLeaderboard(ctx context.Context, req *FetchYearlyLeaderboardRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.r.FetchYearlyLeaderboard(ctx, req)
}
