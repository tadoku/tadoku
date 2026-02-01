package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ProfileYearlyScoresRepository interface {
	YearlyScoresForUser(context.Context, *ProfileYearlyScoresRequest) ([]Score, error)
}

type ProfileYearlyScoresRequest struct {
	UserID uuid.UUID
	Year   int
}

type ProfileYearlyScoresResponse struct {
	OverallScore float32
	Scores       []Score
}

type ProfileYearlyScores struct {
	repo ProfileYearlyScoresRepository
}

func NewProfileYearlyScores(repo ProfileYearlyScoresRepository) *ProfileYearlyScores {
	return &ProfileYearlyScores{repo: repo}
}

func (s *ProfileYearlyScores) Execute(ctx context.Context, req *ProfileYearlyScoresRequest) (*ProfileYearlyScoresResponse, error) {
	scores, err := s.repo.YearlyScoresForUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	response := &ProfileYearlyScoresResponse{
		Scores: scores,
	}

	for _, it := range scores {
		response.OverallScore += it.Score
	}

	return response, nil
}
