package query

import (
	"context"
)

type FetchLeaderboardRequest struct {
	Year         *int16
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

func (s *ServiceImpl) FetchLeaderboard(ctx context.Context, req *FetchLeaderboardRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return s.r.FetchLeaderboard(ctx, req)
}
