package query

import (
	"context"
)

type FetchGlobalLeaderboardRequest struct {
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

func (s *ServiceImpl) FetchGlobalLeaderboard(ctx context.Context, req *FetchGlobalLeaderboardRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.r.FetchGlobalLeaderboard(ctx, req)
}
