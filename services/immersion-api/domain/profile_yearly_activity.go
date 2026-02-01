package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ProfileYearlyActivityRepository interface {
	YearlyActivityForUser(context.Context, *ProfileYearlyActivityRequest) ([]UserActivityScore, error)
}

type ProfileYearlyActivityRequest struct {
	UserID uuid.UUID
	Year   int
}

type ProfileYearlyActivityResponse struct {
	Scores       []UserActivityScore
	TotalUpdates int
}

type ProfileYearlyActivity struct {
	repo ProfileYearlyActivityRepository
}

func NewProfileYearlyActivity(repo ProfileYearlyActivityRepository) *ProfileYearlyActivity {
	return &ProfileYearlyActivity{repo: repo}
}

func (s *ProfileYearlyActivity) Execute(ctx context.Context, req *ProfileYearlyActivityRequest) (*ProfileYearlyActivityResponse, error) {
	scores, err := s.repo.YearlyActivityForUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch activity summary: %w", err)
	}

	res := &ProfileYearlyActivityResponse{
		Scores:       scores,
		TotalUpdates: 0,
	}

	for _, it := range scores {
		res.TotalUpdates += it.Updates
	}

	return res, nil
}
