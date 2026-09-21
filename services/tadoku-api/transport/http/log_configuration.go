package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ImmersionLogGetConfigurations(ctx context.Context, _ openapi.ImmersionLogGetConfigurationsRequestObject) (openapi.ImmersionLogGetConfigurationsResponseObject, error) {
	options, err := s.application.LogConfigurationOptions(ctx)
	if err != nil {
		s.logOperationError(ctx, "get log configuration options", err)
		return nil, err
	}

	response := openapi.ImmersionLogGetConfigurations200JSONResponse{
		Activities:           make([]openapi.ImmersionActivity, 0, len(options.Activities)),
		Languages:            make([]openapi.ImmersionLanguage, 0, len(options.Languages)),
		Units:                make([]openapi.ImmersionUnit, 0, len(options.Units)),
		UserLanguageCodes:    &options.UserLanguageCodes,
		ScoringEngineEnabled: options.ScoringEngineEnabled,
	}
	for _, activity := range options.Activities {
		inputType := openapi.ImmersionActivityInputType(activity.InputType)
		response.Activities = append(response.Activities, openapi.ImmersionActivity{Id: activity.ID, Name: activity.Name, InputType: &inputType})
	}
	for _, language := range options.Languages {
		response.Languages = append(response.Languages, openapi.ImmersionLanguage{Code: language.Code, Name: language.Name})
	}
	for _, unit := range options.Units {
		response.Units = append(response.Units, openapi.ImmersionUnit{
			Id:            unit.ID,
			UnitKey:       unit.Key,
			LogActivityId: unit.LogActivityID,
			Name:          unit.Name,
			Modifier:      unit.Modifier,
			LanguageCode:  unit.LanguageCode,
		})
	}
	return response, nil
}
