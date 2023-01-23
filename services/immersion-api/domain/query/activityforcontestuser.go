package query

import (
	"context"
	"fmt"
	"time"
)

type ActivityForContestUserResponse struct {
	Rows []ActivityForContestUserRow
}

type ActivityForContestUserRow struct {
	Date         time.Time
	LanguageCode string
	Score        float32
}

func (s *ServiceImpl) ActivityForContestUser(ctx context.Context, req *ContestProfileRequest) (*ActivityForContestUserResponse, error) {
	rows, err := s.r.ActivityForContestUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch activity: %w", err)
	}

	return &ActivityForContestUserResponse{
		Rows: rows,
	}, nil
}
