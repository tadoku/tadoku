package rest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Submits a new log
// (POST /logs)
func (s *Server) LogCreate(ctx echo.Context) error {
	var req openapi.LogCreateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	var registrationIDs []uuid.UUID
	if req.RegistrationIds != nil {
		registrationIDs = *req.RegistrationIds
	}

	log, err := s.logCreate.Execute(ctx.Request().Context(), &domain.LogCreateRequest{
		RegistrationIDs: registrationIDs,
		UnitID:          req.UnitId,
		ActivityID:      req.ActivityId,
		LanguageCode:    req.LanguageCode,
		Amount:          req.Amount,
		DurationSeconds: int32PtrFromIntPtr(req.DurationSeconds),
		Tags:            req.Tags,
		Description:     req.Description,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		if errors.Is(err, domain.ErrInvalidLog) {
			ctx.Echo().Logger.Error("could not process request: ", err)
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, logToAPI(log))
}
