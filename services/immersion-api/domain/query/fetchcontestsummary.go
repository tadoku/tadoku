package query

import (
	"context"

	"github.com/google/uuid"
)

type FetchContestSummaryRequest struct {
	ContestID uuid.UUID
}

type FetchContestSummaryResponse struct {
	ParticipantCount int
	LanguageCount    int
	TotalScore       float32
}

func (s *ServiceImpl) FetchContestSummary(ctx context.Context, req *FetchContestSummaryRequest) (*FetchContestSummaryResponse, error) {
	return s.r.FetchContestSummary(ctx, req)
}
