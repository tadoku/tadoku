package query

import (
	"context"
)

type FetchOfficialLeaderboardPreviewForYearRequest struct {
	Year int
}

func (s *ServiceImpl) FetchOfficialLeaderboardPreviewForYear(ctx context.Context, req *FetchOfficialLeaderboardPreviewForYearRequest) (*Leaderboard, error) {
	return s.r.FetchOfficialLeaderboardPreviewForYear(ctx, req)
}
