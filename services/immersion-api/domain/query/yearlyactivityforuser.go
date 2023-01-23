package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type YearlyActivityForUserRequest struct {
	UserID uuid.UUID
	Year   int
}

type YearlyActivityForUserResponse struct {
	Scores       []UserActivityScore
	TotalUpdates int
}

func (s *ServiceImpl) YearlyActivityForUser(ctx context.Context, req *YearlyActivityForUserRequest) (*YearlyActivityForUserResponse, error) {
	scores, err := s.r.YearlyActivityForUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch activity summary: %w", err)
	}

	res := &YearlyActivityForUserResponse{
		Scores:       scores,
		TotalUpdates: 0,
	}

	for _, it := range scores {
		res.TotalUpdates += it.Updates
	}

	return res, nil
}
