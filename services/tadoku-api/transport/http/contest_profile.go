package http

import (
	"context"

	"github.com/oapi-codegen/nullable"
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

func (s *server) ImmersionContestProfileFetchSummary(
	ctx context.Context,
	request openapi.ImmersionContestProfileFetchSummaryRequestObject,
) (openapi.ImmersionContestProfileFetchSummaryResponseObject, error) {
	filters := logFilters(request.Params.Language, request.Params.Activity, request.Params.Tag)
	result, err := s.application.ContestProfileSummary(ctx, request.UserId, request.Id, filters)
	if err != nil {
		s.logOperationError(ctx, "contest profile summary", err)
		return nil, err
	}

	response := openapi.ImmersionContestProfileSummary{
		Score:      result.Score,
		LogCount:   result.LogCount,
		ActiveDays: result.ActiveDays,
		BestDay:    nullable.NewNullNullable[openapi.ImmersionContestProfileSummaryBestDay](),
		Days:       make([]openapi.ImmersionContestProfileSummaryDay, 0, len(result.Days)),
		Languages:  make([]openapi.ImmersionContestProfileSummaryLanguage, 0, len(result.Languages)),
		Activities: make([]openapi.ImmersionContestProfileSummaryActivity, 0, len(result.Activities)),
		Tags:       make([]openapi.ImmersionContestProfileSummaryTag, 0, len(result.Tags)),
	}
	if result.BestDay != nil {
		response.BestDay = nullable.NewNullableWithValue(openapi.ImmersionContestProfileSummaryBestDay{
			Date:  openapiTypes.Date{Time: result.BestDay.Date},
			Score: result.BestDay.Score,
		})
	}
	for _, day := range result.Days {
		response.Days = append(response.Days, openapi.ImmersionContestProfileSummaryDay{
			Date:       openapiTypes.Date{Time: day.Date},
			ActivityId: day.ActivityID,
			Score:      day.Score,
		})
	}
	for _, language := range result.Languages {
		response.Languages = append(response.Languages, openapi.ImmersionContestProfileSummaryLanguage{
			Code:  language.LanguageCode,
			Name:  language.LanguageName,
			Score: language.Score,
		})
	}
	for _, activity := range result.Activities {
		response.Activities = append(response.Activities, openapi.ImmersionContestProfileSummaryActivity{
			Id:    activity.ActivityID,
			Name:  activity.ActivityName,
			Score: activity.Score,
		})
	}
	for _, tag := range result.Tags {
		response.Tags = append(response.Tags, openapi.ImmersionContestProfileSummaryTag{
			Tag:      tag.Tag,
			Score:    tag.Score,
			LogCount: tag.LogCount,
		})
	}
	return openapi.ImmersionContestProfileFetchSummary200JSONResponse(response), nil
}
