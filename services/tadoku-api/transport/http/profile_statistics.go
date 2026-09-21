package http

import (
	"context"

	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionProfileFindByUserID(ctx context.Context, request openapi.ImmersionProfileFindByUserIDRequestObject) (openapi.ImmersionProfileFindByUserIDResponseObject, error) {
	result, err := s.application.FindProfile(ctx, request.UserId)
	if err != nil {
		s.logOperationError(ctx, "FindProfile", err)
		return openapi.ImmersionProfileFindByUserID500Response{}, nil
	}

	return openapi.ImmersionProfileFindByUserID200JSONResponse(openapi.ImmersionUserProfile{
		Id:          request.UserId,
		DisplayName: result.DisplayName,
		CreatedAt:   result.CreatedAt,
	}), nil
}

func (s *server) ImmersionProfileYearlyActivityByUserID(ctx context.Context, request openapi.ImmersionProfileYearlyActivityByUserIDRequestObject) (openapi.ImmersionProfileYearlyActivityByUserIDResponseObject, error) {
	result, err := s.application.YearlyActivity(ctx, request.UserId, request.Year)
	if err != nil {
		s.logOperationError(ctx, "YearlyActivity", err)
		return openapi.ImmersionProfileYearlyActivityByUserID500Response{}, nil
	}

	response := openapi.ImmersionUserActivity{
		TotalUpdates: result.TotalUpdates,
		Scores:       make([]openapi.ImmersionUserActivityScore, 0, len(result.Scores)),
	}
	for _, score := range result.Scores {
		response.Scores = append(response.Scores, openapi.ImmersionUserActivityScore{
			Date:  openapiTypes.Date{Time: score.Date},
			Score: score.Score,
		})
	}

	return openapi.ImmersionProfileYearlyActivityByUserID200JSONResponse(response), nil
}

func (s *server) ImmersionProfileYearlyScoresByUserID(ctx context.Context, request openapi.ImmersionProfileYearlyScoresByUserIDRequestObject) (openapi.ImmersionProfileYearlyScoresByUserIDResponseObject, error) {
	result, err := s.application.YearlyScores(ctx, request.UserId, request.Year)
	if err != nil {
		s.logOperationError(ctx, "YearlyScores", err)
		return openapi.ImmersionProfileYearlyScoresByUserID500Response{}, nil
	}

	response := openapi.ImmersionProfileScores{
		OverallScore: result.OverallScore,
		Scores:       make([]openapi.ImmersionScore, 0, len(result.Scores)),
	}
	for _, score := range result.Scores {
		response.Scores = append(response.Scores, openapi.ImmersionScore{
			LanguageCode: score.LanguageCode,
			LanguageName: &score.LanguageName,
			Score:        score.Score,
		})
	}

	return openapi.ImmersionProfileYearlyScoresByUserID200JSONResponse(response), nil
}

func (s *server) ImmersionProfileYearlyActivitySplitByUserID(ctx context.Context, request openapi.ImmersionProfileYearlyActivitySplitByUserIDRequestObject) (openapi.ImmersionProfileYearlyActivitySplitByUserIDResponseObject, error) {
	result, err := s.application.YearlyActivitySplit(ctx, request.UserId, request.Year)
	if err != nil {
		s.logOperationError(ctx, "YearlyActivitySplit", err)
		return openapi.ImmersionProfileYearlyActivitySplitByUserID500Response{}, nil
	}

	response := openapi.ImmersionActivitySplit{Activities: make([]openapi.ImmersionActivitySplitScore, 0, len(result))}
	for _, score := range result {
		response.Activities = append(response.Activities, openapi.ImmersionActivitySplitScore{
			ActivityId:   score.ActivityID,
			ActivityName: score.ActivityName,
			Score:        score.Score,
		})
	}

	return openapi.ImmersionProfileYearlyActivitySplitByUserID200JSONResponse(response), nil
}
