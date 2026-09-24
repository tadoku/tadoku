package http

import (
	"context"

	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestProfileFetchScores(ctx context.Context, request openapi.ImmersionContestProfileFetchScoresRequestObject) (openapi.ImmersionContestProfileFetchScoresResponseObject, error) {
	result, err := s.application.ContestProfileScores(ctx, request.UserId, request.Id)
	if err != nil {
		s.logOperationError(ctx, "contest profile scores", err)
		return nil, err
	}

	response := openapi.ImmersionContestProfileScores{
		Registration: registrationResponse(result.Registration),
		OverallScore: result.OverallScore,
		Scores:       make([]openapi.ImmersionScore, 0, len(result.Scores)),
	}
	for _, score := range result.Scores {
		response.Scores = append(response.Scores, openapi.ImmersionScore{
			LanguageCode: score.LanguageCode,
			Score:        score.Score,
		})
	}
	return openapi.ImmersionContestProfileFetchScores200JSONResponse(response), nil
}

func (s *server) ImmersionContestProfileFetchActivity(ctx context.Context, request openapi.ImmersionContestProfileFetchActivityRequestObject) (openapi.ImmersionContestProfileFetchActivityResponseObject, error) {
	result, err := s.application.ContestProfileActivity(ctx, request.UserId, request.Id)
	if err != nil {
		s.logOperationError(ctx, "contest profile activity", err)
		return nil, err
	}

	response := openapi.ImmersionContestProfileActivity{Rows: make([]openapi.ImmersionContestProfileActivityRow, 0, len(result))}
	for _, row := range result {
		response.Rows = append(response.Rows, openapi.ImmersionContestProfileActivityRow{
			Date:         openapiTypes.Date{Time: row.Date},
			LanguageCode: row.LanguageCode,
			Score:        row.Score,
		})
	}
	return openapi.ImmersionContestProfileFetchActivity200JSONResponse(response), nil
}
