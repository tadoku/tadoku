package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Fetches a contest registration if it exists
// (GET /contests/{id}/registration)
func (s *Server) ContestFindRegistration(ctx echo.Context, id types.UUID) error {
	reg, err := s.registrationFind.Execute(ctx.Request().Context(), &domain.RegistrationFindRequest{
		ContestID: id,
	})
	if err != nil {
		if handled, respErr := noContentForCommonDomainError(ctx, err); handled {
			return respErr
		}

		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	langs := make([]openapi.Language, len(reg.Languages))
	for i, it := range reg.Languages {
		langs[i] = openapi.Language{
			Code: it.Code,
			Name: it.Name,
		}
	}

	return ctx.JSON(http.StatusOK, openapi.ContestRegistration{
		Id:              &reg.ID,
		ContestId:       reg.ContestID,
		UserId:          reg.UserID,
		UserDisplayName: reg.UserDisplayName,
		Languages:       langs,
	})
}
