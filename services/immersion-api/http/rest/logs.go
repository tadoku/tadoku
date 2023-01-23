package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"

	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
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

	if err := s.commandService.CreateLog(ctx.Request().Context(), &command.LogCreateRequest{
		RegistrationIDs: req.RegistrationIds,
		UnitID:          req.UnitId,
		ActivityID:      req.ActivityId,
		LanguageCode:    req.LanguageCode,
		Amount:          req.Amount,
		Tags:            req.Tags,
		Description:     req.Description,
	}); err != nil {
		if errors.Is(err, command.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, command.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		if errors.Is(err, command.ErrInvalidLog) {
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
	opts, err := s.queryService.FetchLogConfigurationOptions(ctx.Request().Context())
	if err != nil {
		if errors.Is(err, query.ErrUnauthorized) {
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
	log, err := s.queryService.FindLogByID(ctx.Request().Context(), &query.FindLogByIDRequest{
		ID: id,
	})
	if err != nil {
		if errors.Is(err, query.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	refs := make([]openapi.ContestRegistrationReference, len(log.Registrations))
	for i, it := range log.Registrations {
		refs[i] = openapi.ContestRegistrationReference{
			ContestId:      it.ContestID,
			RegistrationId: it.RegistrationID,
			Title:          it.Title,
		}
	}

	res := openapi.Log{
		Id: log.ID,
		Activity: openapi.Activity{
			Id:   int32(log.ActivityID),
			Name: log.ActivityName,
		},
		Language: openapi.Language{
			Code: log.LanguageCode,
			Name: log.LanguageName,
		},
		Amount:          log.Amount,
		Modifier:        log.Modifier,
		Score:           log.Score,
		Tags:            log.Tags,
		UnitName:        log.UnitName,
		UserId:          log.UserID,
		UserDisplayName: log.UserDisplayName,
		CreatedAt:       log.CreatedAt,
		Deleted:         log.Deleted,
		Description:     log.Description,
		Registrations:   &refs,
	}

	return ctx.JSON(http.StatusOK, res)
}
