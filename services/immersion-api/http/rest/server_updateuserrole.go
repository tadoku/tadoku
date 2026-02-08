package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/http/rest/openapi"
)

// Update user role (admin only)
// (PUT /users/{id}/role)
func (s *Server) UpdateUserRole(ctx echo.Context, id types.UUID) error {
	var req openapi.UpdateUserRoleJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	err := s.updateUserRole.Execute(ctx.Request().Context(), &domain.UpdateUserRoleRequest{
		UserID: id,
		Role:   string(req.Role),
		Reason: req.Reason,
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
