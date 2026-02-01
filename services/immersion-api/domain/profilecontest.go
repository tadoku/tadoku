package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ProfileContestRepository interface {
	FindRegistrationForUser(context.Context, *RegistrationFindRequest) (*ContestRegistration, error)
	FindScoresForRegistration(context.Context, *ProfileContestRequest) ([]Score, error)
}

type ProfileContestRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type ProfileContestResponse struct {
	Registration *ContestRegistration
	OverallScore float32
	Scores       []Score
}

type ProfileContest struct {
	repo ProfileContestRepository
}

func NewProfileContest(repo ProfileContestRepository) *ProfileContest {
	return &ProfileContest{repo: repo}
}

func (s *ProfileContest) Execute(ctx context.Context, req *ProfileContestRequest) (*ProfileContestResponse, error) {
	reg, err := s.repo.FindRegistrationForUser(ctx, &RegistrationFindRequest{
		UserID:    req.UserID,
		ContestID: req.ContestID,
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch registration: %w", err)
	}

	response := &ProfileContestResponse{
		Registration: reg,
	}

	scores, err := s.repo.FindScoresForRegistration(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	response.Scores = scores

	for _, it := range scores {
		response.OverallScore += it.Score
	}

	return response, nil
}
