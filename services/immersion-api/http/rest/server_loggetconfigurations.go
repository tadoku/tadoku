package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the configuration options for a log
// (GET /logs/configuration-options)
func (s *Server) LogGetConfigurations(ctx echo.Context) error {
	opts, err := s.logConfigurationOptions.Execute(ctx.Request().Context())
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		return ctx.NoContent(http.StatusInternalServerError)
	}

	userLangCodes := opts.UserLanguageCodes
	if userLangCodes == nil {
		userLangCodes = []string{}
	}

	res := openapi.LogConfigurationOptions{
		Activities:        make([]openapi.Activity, len(opts.Activities)),
		Languages:         make([]openapi.Language, len(opts.Languages)),
		Units:             make([]openapi.Unit, len(opts.Units)),
		UserLanguageCodes: &userLangCodes,
	}

	for i, it := range opts.Activities {
		it := it
		res.Activities[i] = openapi.Activity{
			Id:           it.ID,
			Name:         it.Name,
			InputType:    openapi.ActivityInputType(it.InputType),
			TimeModifier: &it.TimeModifier,
		}
	}

	for i, it := range opts.Languages {
		res.Languages[i] = openapi.Language{
			Code: it.Code,
			Name: it.Name,
		}
	}

	for i, it := range opts.Units {
		res.Units[i] = openapi.Unit{
			Id:            it.ID,
			LogActivityId: it.LogActivityID,
			Name:          it.Name,
			Modifier:      it.Modifier,
			LanguageCode:  it.LanguageCode,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}
