package http

import (
	"context"
	"errors"

	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestCreate(
	ctx context.Context,
	request openapi.ImmersionContestCreateRequestObject,
) (openapi.ImmersionContestCreateResponseObject, error) {
	parameters := app.CreateContestParameters{
		ContestStart:            request.Body.ContestStart.Time,
		ContestEnd:              request.Body.ContestEnd.Time,
		RegistrationEnd:         request.Body.RegistrationEnd.Time,
		Title:                   request.Body.Title,
		Description:             request.Body.Description,
		Official:                request.Body.Official,
		Private:                 request.Body.Private,
		LanguageCodeAllowList:   request.Body.LanguageCodeAllowList,
		ActivityTypeIDAllowList: request.Body.ActivityTypeIdAllowList,
	}
	contest, err := s.application.CreateContest(ctx, parameters)
	if err != nil {
		s.logOperationError(ctx, "create contest", err)
		if errors.Is(err, app.ErrAccountDeletionInProgress) {
			return openapi.ImmersionContestCreate409JSONResponse{
				ImmersionAccountDeletionInProgressJSONResponse: openapi.ImmersionAccountDeletionInProgressJSONResponse{
					Error: openapi.AccountDeletionInProgress,
				},
			}, nil
		}
		if errors.Is(err, app.ErrInvalidContestCreator) {
			return openapi.ImmersionContestCreate500JSONResponse{Message: "Internal Server Error"}, nil
		}
		return nil, err
	}

	return openapi.ImmersionContestCreate200JSONResponse(contestResponse(contest)), nil
}

func contestResponse(item *app.Contest) openapi.ImmersionContest {
	return openapi.ImmersionContest{
		ActivityTypeIdAllowList: item.ActivityTypeIDAllowList,
		ContestEnd:              openapiTypes.Date{Time: item.ContestEnd},
		ContestStart:            openapiTypes.Date{Time: item.ContestStart},
		CreatedAt:               &item.CreatedAt,
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
	}
}
