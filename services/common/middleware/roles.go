package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/authz/roles"
	"github.com/tadoku/tadoku/services/common/domain"
)

// RolesFromKeto enriches request context with role claims for authenticated users.
// It does not block the request on evaluation failures; downstream code should use
// roles.RequireAdmin / roles.IsBanned etc which will surface ErrAuthzUnavailable.
func RolesFromKeto(svc roles.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			user := domain.ParseUserIdentity(ctx.Request().Context())
			if user == nil || user.Subject == "" || user.Subject == "guest" {
				return next(ctx)
			}

			claims, err := svc.ClaimsForSubject(ctx.Request().Context(), user.Subject)
			if err != nil {
				claims.Err = err
			}

			ctx.SetRequest(ctx.Request().WithContext(roles.WithClaims(ctx.Request().Context(), claims)))

			return next(ctx)
		}
	}
}
