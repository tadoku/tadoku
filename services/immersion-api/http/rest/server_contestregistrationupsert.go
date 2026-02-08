package rest

import (
	"errors"
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Creates or updates a registration for a contest
// (POST /contests/{id}/registration)
func (s *Server) ContestRegistrationUpsert(ctx echo.Context, id types.UUID) error {
	var req openapi.ContestRegistrationUpsertJSONBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err := s.registrationUpsert.Execute(ctx.Request().Context(), &domain.RegistrationUpsertRequest{
		ContestID:     id,
		LanguageCodes: req.LanguageCodes,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		if errors.Is(err, domain.ErrInvalidContestRegistration) {
			return ctx.NoContent(http.StatusBadRequest)
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
