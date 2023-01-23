package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type YearlyScoresForUserRequest struct {
	UserID uuid.UUID
	Year   int
}

type YearlyScoresForUserResponse struct {
	OverallScore float32
	Scores       []Score
}

func (s *ServiceImpl) YearlyScoresForUser(ctx context.Context, req *YearlyScoresForUserRequest) (*YearlyScoresForUserResponse, error) {
	scores, err := s.r.YearlyScoresForUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	response := &YearlyScoresForUserResponse{
		Scores: scores,
	}

	for _, it := range scores {
		response.OverallScore += it.Score
	}

	return response, nil
}
