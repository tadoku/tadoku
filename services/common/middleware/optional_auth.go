package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/tadoku/tadoku/services/common/authz/roles"
)

// OptionalAdminAuth creates middleware that allows unauthenticated requests
// (e.g. Kubernetes probes) but requires admin authorization when a JWT
// bearer token is present (e.g. external requests through the API gateway).
func OptionalAdminAuth(jwksURL string, rolesSvc roles.Service) echo.MiddlewareFunc {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		panic(fmt.Errorf("optional auth: unable to fetch jwks: %w", err))
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				return next(c)
			}

			tokenString := strings.TrimPrefix(auth, "Bearer ")
			token, err := jwtv4.ParseWithClaims(tokenString, &UnifiedClaims{}, jwks.Keyfunc)
			if err != nil || !token.Valid {
				return c.NoContent(http.StatusUnauthorized)
			}

			claims, ok := token.Claims.(*UnifiedClaims)
			if !ok {
				return c.NoContent(http.StatusUnauthorized)
			}

			// Service tokens with valid JWT are allowed (internal service-to-service)
			if claims.Type == "service" {
				return next(c)
			}

			// User tokens must be admin
			subject := claims.Subject
			if subject == "" || subject == "guest" {
				return c.NoContent(http.StatusUnauthorized)
			}

			roleClaims, err := rolesSvc.ClaimsForSubject(c.Request().Context(), subject)
			if err != nil {
				return c.NoContent(http.StatusServiceUnavailable)
			}
			if !roleClaims.Admin {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}
