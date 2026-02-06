package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Updates an existing language (admin only)
// (PUT /languages/{code})
func (s *Server) LanguageUpdate(ctx echo.Context, code string) error {
	var req openapi.LanguageUpdateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err := s.languageUpdate.Execute(ctx.Request().Context(), &domain.LanguageUpdateRequest{
		Code: code,
		Name: req.Name,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		if errors.Is(err, domain.ErrRequestInvalid) {
			return ctx.NoContent(http.StatusBadRequest)
		}
		if errors.Is(err, domain.ErrNotFound) {
			return ctx.NoContent(http.StatusNotFound)
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
