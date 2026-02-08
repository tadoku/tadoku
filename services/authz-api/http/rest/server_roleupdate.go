package rest

import (
	"net/http"

	"github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi"
)

// (PUT /users/{id}/role)
func (s *Server) RoleUpdate(ctx echo.Context, id types.UUID) error {
	var req openapi.RoleUpdateJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	userID, err := uuid.Parse(id.String())
	if err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	err = s.roleUpdate.Execute(ctx.Request().Context(), &domain.RoleUpdateRequest{
		UserID: userID,
		Role:   string(req.Role),
		Reason: req.Reason,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}

