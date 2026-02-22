package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches the configuration options for a new contest
// (GET /contests/configuration-options)
func (s *Server) ContestGetConfigurations(ctx echo.Context) error {
	opts, err := s.contestConfigurationOptions.Execute(ctx.Request().Context())
	if err != nil {
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.ContestConfigurationOptions{
		Activities:             make([]openapi.Activity, len(opts.Activities)),
		Languages:              make([]openapi.Language, len(opts.Languages)),
		CanCreateOfficialRound: opts.CanCreateOfficialRound,
	}

	for i, a := range opts.Activities {
		a := a
		res.Activities[i] = openapi.Activity{
			Id:        a.ID,
			Name:      a.Name,
			Default:   &a.Default,
			InputType: openapi.ActivityInputType(a.InputType),
		}
	}

	for i, l := range opts.Languages {
		res.Languages[i] = openapi.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}
