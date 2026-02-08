package rest

import (
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
		if handled, respErr := handleCommonDomainError(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}
