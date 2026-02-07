package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Creates a new language (admin only)
// (POST /languages)
func (s *Server) LanguageCreate(ctx echo.Context) error {
	var req openapi.LanguageCreateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err := s.languageCreate.Execute(ctx.Request().Context(), &domain.LanguageCreateRequest{
		Code: req.Code,
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
		if errors.Is(err, domain.ErrConflict) {
			return ctx.NoContent(http.StatusConflict)
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
