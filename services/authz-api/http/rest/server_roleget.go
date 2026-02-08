package rest

import (
	"net/http"

	"github.com/labstack/echo/v4"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// (GET /current-user/role)
func (s *Server) RoleGet(ctx echo.Context) error {
	session := commondomain.ParseUserIdentity(ctx.Request().Context())
	if session == nil {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	role, err := s.roleGet.Execute(ctx.Request().Context(), session.Subject)
	if err != nil {
		if handled, respErr := handleCommonErrors(ctx, err); handled {
			return respErr
		}
		ctx.Echo().Logger.Error(err)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"role": role})
}

