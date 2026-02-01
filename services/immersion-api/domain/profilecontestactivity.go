package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ProfileContestActivityRepository interface {
	ActivityForContestUser(context.Context, *ProfileContestActivityRequest) ([]ProfileContestActivityRow, error)
}

type ProfileContestActivityRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type ProfileContestActivityResponse struct {
	Rows []ProfileContestActivityRow
}

type ProfileContestActivityRow struct {
	Date         time.Time
	LanguageCode string
	Score        float32
}

type ProfileContestActivity struct {
	repo ProfileContestActivityRepository
}

func NewProfileContestActivity(repo ProfileContestActivityRepository) *ProfileContestActivity {
	return &ProfileContestActivity{repo: repo}
}

func (s *ProfileContestActivity) Execute(ctx context.Context, req *ProfileContestActivityRequest) (*ProfileContestActivityResponse, error) {
	rows, err := s.repo.ActivityForContestUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch activity: %w", err)
	}

	return &ProfileContestActivityResponse{
		Rows: rows,
	}, nil
}
