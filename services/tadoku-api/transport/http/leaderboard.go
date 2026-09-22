package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (s *server) ImmersionContestFetchLeaderboard(ctx context.Context, request openapi.ImmersionContestFetchLeaderboardRequestObject) (openapi.ImmersionContestFetchLeaderboardResponseObject, error) {
	result, err := s.application.FetchContestLeaderboard(ctx, app.ContestLeaderboardRequest{
		ContestID: request.Id,
		Request:   requestFromParams(request.Params.PageSize, request.Params.Page, request.Params.LanguageCode, request.Params.ActivityId),
	})
	if err != nil {
		s.logOperationError(ctx, "fetch contest leaderboard", err)
		switch errx.KindOf(err) {
		case errx.InvalidInput, errx.NotFound:
			return nil, err
		default:
			return openapi.ImmersionContestFetchLeaderboard500Response{}, nil
		}
	}
	return openapi.ImmersionContestFetchLeaderboard200JSONResponse(responseLeaderboard(result)), nil
}

func (s *server) ImmersionFetchLeaderboardForYear(ctx context.Context, request openapi.ImmersionFetchLeaderboardForYearRequestObject) (openapi.ImmersionFetchLeaderboardForYearResponseObject, error) {
	result, err := s.application.FetchYearlyLeaderboard(ctx, app.YearlyLeaderboardRequest{
		Year:    int32(request.Year),
		Request: requestFromParams(request.Params.PageSize, request.Params.Page, request.Params.LanguageCode, request.Params.ActivityId),
	})
	if err != nil {
		s.logOperationError(ctx, "fetch yearly leaderboard", err)
		if errx.KindOf(err) == errx.InvalidInput {
			return nil, err
		}
		return openapi.ImmersionFetchLeaderboardForYear500Response{}, nil
	}
	return openapi.ImmersionFetchLeaderboardForYear200JSONResponse(responseLeaderboard(result)), nil
}

func (s *server) ImmersionFetchLeaderboardGlobal(ctx context.Context, request openapi.ImmersionFetchLeaderboardGlobalRequestObject) (openapi.ImmersionFetchLeaderboardGlobalResponseObject, error) {
	result, err := s.application.FetchGlobalLeaderboard(ctx, requestFromParams(request.Params.PageSize, request.Params.Page, request.Params.LanguageCode, request.Params.ActivityId))
	if err != nil {
		s.logOperationError(ctx, "fetch global leaderboard", err)
		if errx.KindOf(err) == errx.InvalidInput {
			return nil, err
		}
		return openapi.ImmersionFetchLeaderboardGlobal500Response{}, nil
	}
	return openapi.ImmersionFetchLeaderboardGlobal200JSONResponse(responseLeaderboard(result)), nil
}

func requestFromParams(pageSize, page *int, languageCode *string, activityID *int) app.LeaderboardRequest {
	request := app.LeaderboardRequest{LanguageCode: languageCode}
	if pageSize != nil {
		request.PageSize = *pageSize
	}
	if page != nil {
		request.Page = *page
	}
	if activityID != nil {
		narrowed := int32(*activityID)
		request.ActivityID = &narrowed
	}
	return request
}

func responseLeaderboard(result *app.Leaderboard) openapi.ImmersionLeaderboard {
	response := openapi.ImmersionLeaderboard{
		Entries:       make([]openapi.ImmersionLeaderboardEntry, len(result.Entries)),
		TotalSize:     result.TotalSize,
		NextPageToken: result.NextPageToken,
	}
	for i, entry := range result.Entries {
		response.Entries[i] = openapi.ImmersionLeaderboardEntry{
			Rank:            entry.Rank,
			UserId:          entry.UserID,
			UserDisplayName: entry.UserDisplayName,
			Score:           entry.Score,
			IsTie:           entry.IsTie,
		}
	}
	return response
}
