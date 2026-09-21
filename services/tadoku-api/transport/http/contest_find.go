package http

import (
	"context"

	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestFindByID(
	ctx context.Context,
	request openapi.ImmersionContestFindByIDRequestObject,
) (openapi.ImmersionContestFindByIDResponseObject, error) {
	item, err := s.application.FindContestByID(ctx, request.Id)
	if err != nil {
		s.logOperationError(ctx, "find contest by ID", err)
		return nil, err
	}
	return openapi.ImmersionContestFindByID200JSONResponse(contestViewResponse(item)), nil
}

func (s *server) ImmersionContestFindLatestOfficial(
	ctx context.Context,
	_ openapi.ImmersionContestFindLatestOfficialRequestObject,
) (openapi.ImmersionContestFindLatestOfficialResponseObject, error) {
	item, err := s.application.FindLatestOfficialContest(ctx)
	if err != nil {
		s.logOperationError(ctx, "find latest official contest", err)
		return nil, err
	}
	return openapi.ImmersionContestFindLatestOfficial200JSONResponse(contestViewResponse(item)), nil
}

func contestViewResponse(item *app.ContestView) openapi.ImmersionContestView {
	activities := make([]openapi.ImmersionActivity, 0, len(item.AllowedActivities))
	for _, activity := range item.AllowedActivities {
		inputType := openapi.ImmersionActivityInputType(activity.InputType)
		activities = append(activities, openapi.ImmersionActivity{
			Id:        activity.ID,
			InputType: &inputType,
			Name:      activity.Name,
		})
	}
	languages := make([]openapi.ImmersionLanguage, 0, len(item.AllowedLanguages))
	for _, language := range item.AllowedLanguages {
		languages = append(languages, openapi.ImmersionLanguage{Code: language.Code, Name: language.Name})
	}
	if len(languages) == 0 {
		languages = nil
	}
	deleted := item.Deleted
	return openapi.ImmersionContestView{
		AllowedActivities:    activities,
		AllowedLanguages:     languages,
		ContestEnd:           openapiTypes.Date{Time: item.ContestEnd},
		ContestStart:         openapiTypes.Date{Time: item.ContestStart},
		CreatedAt:            &item.CreatedAt,
		Deleted:              &deleted,
		Description:          item.Description,
		Id:                   &item.ID,
		Official:             item.Official,
		OwnerUserDisplayName: &item.OwnerUserDisplayName,
		OwnerUserId:          &item.OwnerUserID,
		Private:              item.Private,
		RegistrationEnd:      openapiTypes.Date{Time: item.RegistrationEnd},
		Title:                item.Title,
		UpdatedAt:            &item.UpdatedAt,
	}
}
