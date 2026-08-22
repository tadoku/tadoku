package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/authz-api/http/rest/openapi/proxyapi"
)

// (POST /internal/v1/proxy/admin-check)
func (s *Server) ProxyAdminCheck(ctx echo.Context) error {
	var req proxyapi.ProxyAdminCheckJSONRequestBody
	if err := ctx.Bind(&req); err != nil {
		ctx.Echo().Logger.Error("could not process request: ", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	allowed, err := s.proxyAdminCheck.Execute(ctx.Request().Context(), req.Subject.String())
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}
	if !allowed {
		return ctx.NoContent(http.StatusForbidden)
	}

	// Oathkeeper's remote_json authorizer accepts exactly 200 as an allow decision.
	return ctx.NoContent(http.StatusOK)
}
