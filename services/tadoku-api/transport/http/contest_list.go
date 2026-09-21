package http

import (
	"context"

	openapiTypes "github.com/oapi-codegen/runtime/types"
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
	for _, item := range result.Contests {
		deleted := item.Deleted
		response.Contests = append(response.Contests, openapi.ImmersionContest{
			ActivityTypeIdAllowList: item.ActivityTypeIDAllowList,
			ContestEnd:              openapiTypes.Date{Time: item.ContestEnd},
			ContestStart:            openapiTypes.Date{Time: item.ContestStart},
			CreatedAt:               &item.CreatedAt,
			Deleted:                 &deleted,
			Description:             item.Description,
			Id:                      &item.ID,
			LanguageCodeAllowList:   item.LanguageCodeAllowList,
			Official:                item.Official,
			OwnerUserDisplayName:    &item.OwnerUserDisplayName,
			OwnerUserId:             &item.OwnerUserID,
			Private:                 item.Private,
			RegistrationEnd:         openapiTypes.Date{Time: item.RegistrationEnd},
			Title:                   item.Title,
			UpdatedAt:               &item.UpdatedAt,
		})
	}
	return response, nil
}
