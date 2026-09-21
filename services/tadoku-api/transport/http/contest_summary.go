package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestFetchSummary(
	ctx context.Context,
	request openapi.ImmersionContestFetchSummaryRequestObject,
) (openapi.ImmersionContestFetchSummaryResponseObject, error) {
	summary, err := s.application.FetchContestSummary(ctx, request.Id)
	if err != nil {
		s.logOperationError(ctx, "fetch contest summary", err)
		return nil, err
	}

	return openapi.ImmersionContestFetchSummary200JSONResponse{
		ParticipantCount: summary.ParticipantCount,
		LanguageCount:    summary.LanguageCount,
		TotalScore:       summary.TotalScore,
	}, nil
}
