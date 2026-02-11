package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/authz-api/domain"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi"
)

// (POST /permission/check)
func (s *Server) PermissionCheck(ctx echo.Context) error {
	var req openapi.PermissionCheckJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	allowed, err := s.publicPermissionCheck.Execute(ctx.Request().Context(), domain.PermissionCheckRequest{
		Namespace: req.Namespace,
		Object:    req.Object,
		Relation:  req.Relation,
	})
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, openapi.PermissionCheckResponse{Allowed: allowed})
}
