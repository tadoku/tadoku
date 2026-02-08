package rest

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/authz-api/domain"
)

// RequireInternalServiceAuth blocks requests to /internal/* unless they come from an allowlisted service identity.
func RequireInternalServiceAuth(allowlist domain.ServiceAllowlist) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if !strings.HasPrefix(ctx.Request().URL.Path, "/internal/") {
				return next(ctx)
			}

			svc := commondomain.ParseServiceIdentity(ctx.Request().Context())
			if svc == nil {
				return ctx.NoContent(http.StatusUnauthorized)
			}
			if !allowlist.Allows(svc.Name) {
				return ctx.NoContent(http.StatusForbidden)
			}

			return next(ctx)
		}
	}
}

