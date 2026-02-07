package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Lists all languages (admin only)
// (GET /languages)
func (s *Server) LanguageList(ctx echo.Context) error {
	languages, err := s.languageList.Execute(ctx.Request().Context())
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return ctx.NoContent(http.StatusUnauthorized)
		}
		if errors.Is(err, domain.ErrForbidden) {
			return ctx.NoContent(http.StatusForbidden)
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	result := make([]openapi.Language, len(languages))
	for i, l := range languages {
		result[i] = openapi.Language{
			Code: l.Code,
			Name: l.Name,
		}
	}

	return ctx.JSON(http.StatusOK, openapi.Languages{
		Languages: result,
	})
}
