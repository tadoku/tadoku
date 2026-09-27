package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestList(
	ctx context.Context,
	request openapi.ImmersionContestListRequestObject,
) (openapi.ImmersionContestListResponseObject, error) {
	parameters := app.ListContestsParameters{Official: true}
	if request.Params.PageSize != nil {
		parameters.PageSize = *request.Params.PageSize
	}
	if request.Params.Page != nil {
		parameters.Page = *request.Params.Page
	}
	if request.Params.IncludeDeleted != nil {
		parameters.IncludeDeleted = *request.Params.IncludeDeleted
	}
	if request.Params.Official != nil {
		parameters.Official = *request.Params.Official
	}
	if request.Params.UserId != nil {
		parameters.UserID = request.Params.UserId
	}

	result, err := s.application.ListContests(ctx, parameters)
	if err != nil {
		s.logOperationError(ctx, "list contests", err)
		return nil, err
	}

	response := openapi.ImmersionContestList200JSONResponse{
		Contests:      make([]openapi.ImmersionContest, 0, len(result.Contests)),
		NextPageToken: result.NextPageToken,
		TotalSize:     result.TotalSize,
	}
	for i := range result.Contests {
		response.Contests = append(response.Contests, contestResponse(&result.Contests[i]))
	}
	return response, nil
}
