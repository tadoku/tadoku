package http

import (
	"context"
	"errors"

	openapiTypes "github.com/oapi-codegen/runtime/types"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestFindRegistration(
	ctx context.Context,
	request openapi.ImmersionContestFindRegistrationRequestObject,
) (openapi.ImmersionContestFindRegistrationResponseObject, error) {
	registration, err := s.application.FindContestRegistration(ctx, request.Id)
	if err != nil {
		s.logOperationError(ctx, "find contest registration", err)
		switch {
		case errors.Is(err, app.ErrContestRegistrationNotFound):
			return openapi.ImmersionContestFindRegistration204Response{}, nil
		default:
			return nil, err
		}
	}
	return openapi.ImmersionContestFindRegistration200JSONResponse(registrationResponse(registration)), nil
}

func (s *server) ImmersionContestFindOngoingRegistrations(
	ctx context.Context,
	_ openapi.ImmersionContestFindOngoingRegistrationsRequestObject,
) (openapi.ImmersionContestFindOngoingRegistrationsResponseObject, error) {
	registrations, err := s.application.ListOngoingContestRegistrations(ctx)
	if err != nil {
		s.logOperationError(ctx, "list ongoing contest registrations", err)
		return nil, err
	}
	response := openapi.ImmersionContestFindOngoingRegistrations200JSONResponse{
		Registrations: make([]openapi.ImmersionContestRegistration, 0, len(registrations.Registrations)),
		TotalSize:     registrations.TotalSize,
		NextPageToken: registrations.NextPageToken,
	}
	for i := range registrations.Registrations {
		response.Registrations = append(response.Registrations, registrationResponse(&registrations.Registrations[i]))
	}
	return response, nil
}

func (s *server) ImmersionContestRegistrationUpsert(
	ctx context.Context,
	request openapi.ImmersionContestRegistrationUpsertRequestObject,
) (openapi.ImmersionContestRegistrationUpsertResponseObject, error) {
	var languageCodes []string
	if request.Body != nil {
		languageCodes = request.Body.LanguageCodes
	}
	err := s.application.UpsertContestRegistration(ctx, app.ContestRegistrationUpsertParameters{
		ContestID:     request.Id,
		LanguageCodes: languageCodes,
	})
	if err != nil {
		s.logOperationError(ctx, "upsert contest registration", err)
		return nil, err
	}
	return openapi.ImmersionContestRegistrationUpsert200Response{}, nil
}

func registrationResponse(registration *app.ContestRegistration) openapi.ImmersionContestRegistration {
	response := openapi.ImmersionContestRegistration{
		ContestId:       registration.ContestID,
		Id:              &registration.ID,
		Languages:       make([]openapi.ImmersionLanguage, 0, len(registration.Languages)),
		UserDisplayName: registration.UserDisplayName,
		UserId:          registration.UserID,
	}
	for _, language := range registration.Languages {
		response.Languages = append(response.Languages, openapi.ImmersionLanguage{
			Code: language.Code,
			Name: language.Name,
		})
	}
	if registration.Contest == nil {
		return response
	}

	contest := registration.Contest
	response.Contest = &openapi.ImmersionContestView{
		AllowedActivities:    make([]openapi.ImmersionActivity, 0, len(contest.AllowedActivities)),
		AllowedLanguages:     []openapi.ImmersionLanguage{},
		ContestEnd:           openapiTypes.Date{Time: contest.ContestEnd},
		ContestStart:         openapiTypes.Date{Time: contest.ContestStart},
		Description:          contest.Description,
		Id:                   &contest.ID,
		Official:             contest.Official,
		OwnerUserDisplayName: &contest.OwnerUserDisplayName,
		OwnerUserId:          &contest.OwnerUserID,
		Private:              contest.Private,
		RegistrationEnd:      openapiTypes.Date{Time: contest.RegistrationEnd},
		Title:                contest.Title,
	}
	for _, activity := range contest.AllowedActivities {
		inputType := openapi.ImmersionActivityInputType(activity.InputType)
		response.Contest.AllowedActivities = append(response.Contest.AllowedActivities, openapi.ImmersionActivity{
			Id:        activity.ID,
			InputType: &inputType,
			Name:      activity.Name,
		})
	}
	return response
}
