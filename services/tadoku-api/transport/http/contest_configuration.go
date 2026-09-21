package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionContestGetConfigurations(
	ctx context.Context,
	_ openapi.ImmersionContestGetConfigurationsRequestObject,
) (openapi.ImmersionContestGetConfigurationsResponseObject, error) {
	options, err := s.application.ContestConfigurationOptions(ctx)
	if err != nil {
		s.logOperationError(ctx, "get contest configuration options", err)
		return nil, err
	}

	response := openapi.ImmersionContestGetConfigurations200JSONResponse{
		Activities:             make([]openapi.ImmersionActivity, 0, len(options.Activities)),
		CanCreateOfficialRound: options.CanCreateOfficialRound,
		Languages:              make([]openapi.ImmersionLanguage, 0, len(options.Languages)),
	}
	for _, activity := range options.Activities {
		defaultActivity := activity.Default
		inputType := openapi.ImmersionActivityInputType(activity.InputType)
		response.Activities = append(response.Activities, openapi.ImmersionActivity{
			Default:   &defaultActivity,
			Id:        activity.ID,
			InputType: &inputType,
			Name:      activity.Name,
		})
	}
	for _, language := range options.Languages {
		response.Languages = append(response.Languages, openapi.ImmersionLanguage{
			Code: language.Code,
			Name: language.Name,
		})
	}
	return response, nil
}
