package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"

	"github.com/tadoku/tadoku/services/immersion-api/domain/logcommand"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logquery"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// COMMANDS

// Submits a new log
// (POST /logs)
func (s *Server) LogCreateLog(ctx echo.Context) error {
	var req openapi.LogCreateLogJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := s.logCommandService.CreateLog(ctx.Request().Context(), &logcommand.LogCreateRequest{
		RegistrationIDs: req.RegistrationIds,
		UnitID:          req.UnitId,
		ActivityID:      req.ActivityId,
		LanguageCode:    req.LanguageCode,
		Amount:          req.Amount,
		Tags:            req.Tags,
		Description:     req.Description,
	}); err != nil {
		if errors.Is(err, logcommand.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, logcommand.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		if errors.Is(err, logcommand.ErrInvalidLog) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}

// QUERIES

// Fetches the configuration options for a log
// (GET /logs/configuration-options)
func (s *Server) LogGetConfigurations(ctx echo.Context) error {
	opts, err := s.logQueryService.FetchLogConfigurationOptions(ctx.Request().Context())
	if err != nil {
		if errors.Is(err, logquery.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		return ctx.NoContent(http.StatusInternalServerError)
	}

	res := openapi.LogConfigurationOptions{
		Activities: make([]openapi.Activity, len(opts.Activities)),
		Languages:  make([]openapi.Language, len(opts.Languages)),
		Units:      make([]openapi.Unit, len(opts.Units)),
		Tags:       make([]openapi.Tag, len(opts.Tags)),
	}

	for i, it := range opts.Activities {
		it := it
		res.Activities[i] = openapi.Activity{
			Id:   it.ID,
			Name: it.Name,
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

	for i, it := range opts.Tags {
		res.Tags[i] = openapi.Tag{
			Id:            it.ID,
			LogActivityId: it.LogActivityID,
			Name:          it.Name,
		}
	}

	return ctx.JSON(http.StatusOK, res)
}

// Fetches a log by id
// (GET /logs/{id})
func (s *Server) LogFindByID(ctx echo.Context, id types.UUID) error {
	return nil
}
