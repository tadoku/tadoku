package rest

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func OathkeeperAuthorization(expectedToken string) echo.MiddlewareFunc {
	expectedHash := sha256.Sum256([]byte(expectedToken))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			scheme, token, found := strings.Cut(ctx.Request().Header.Get(echo.HeaderAuthorization), " ")
			providedHash := sha256.Sum256([]byte(token))
			if expectedToken == "" || !found || !strings.EqualFold(scheme, "Bearer") || token == "" ||
				subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) != 1 {
				return ctx.NoContent(http.StatusUnauthorized)
			}

			return next(ctx)
		}
	}
}
