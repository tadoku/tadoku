package domain

import (
	"context"

	"github.com/google/uuid"
)

type ContestSummaryFetchRepository interface {
	FetchContestSummary(ctx context.Context, req *ContestSummaryFetchRequest) (*ContestSummaryFetchResponse, error)
}

type ContestSummaryFetchRequest struct {
	ContestID uuid.UUID
}

type ContestSummaryFetchResponse struct {
	ParticipantCount int
	LanguageCount    int
	TotalScore       float32
}

type ContestSummaryFetch struct {
	repo ContestSummaryFetchRepository
}

func NewContestSummaryFetch(repo ContestSummaryFetchRepository) *ContestSummaryFetch {
	return &ContestSummaryFetch{repo: repo}
}

func (s *ContestSummaryFetch) Execute(ctx context.Context, req *ContestSummaryFetchRequest) (*ContestSummaryFetchResponse, error) {
	return s.repo.FetchContestSummary(ctx, req)
}
