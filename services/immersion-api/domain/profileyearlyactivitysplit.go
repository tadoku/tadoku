package domain

import (
	"context"

	"github.com/google/uuid"
)

type ProfileYearlyActivitySplitRepository interface {
	YearlyActivitySplitForUser(ctx context.Context, req *ProfileYearlyActivitySplitRequest) (*ProfileYearlyActivitySplitResponse, error)
}

type ProfileYearlyActivitySplitRequest struct {
	UserID uuid.UUID
	Year   int
}

type ProfileYearlyActivitySplitResponse struct {
	Activities []ActivityScore
}

type ProfileYearlyActivitySplit struct {
	repo ProfileYearlyActivitySplitRepository
}

func NewProfileYearlyActivitySplit(repo ProfileYearlyActivitySplitRepository) *ProfileYearlyActivitySplit {
	return &ProfileYearlyActivitySplit{repo: repo}
}

func (s *ProfileYearlyActivitySplit) Execute(ctx context.Context, req *ProfileYearlyActivitySplitRequest) (*ProfileYearlyActivitySplitResponse, error) {
	res, err := s.repo.YearlyActivitySplitForUser(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := hydrateActivityScores(res.Activities); err != nil {
		return nil, err
	}
	return res, nil
}
