package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tadoku/tadoku/services/common/domain"
)

func SessionJWT(jwksURL string) echo.MiddlewareFunc {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			panic(fmt.Errorf("unable to refresh jwks: %w", err))
		},
	})

	if err != nil {
		panic(fmt.Errorf("unable to fetch jwks: %w", err))
	}

	return middleware.JWTWithConfig(middleware.JWTConfig{
		Skipper: func(context echo.Context) bool {
			path := context.Path()
			return path == "/ping" || strings.HasPrefix(path, "/internal/")
		},
		Claims: &SessionClaims{},
		KeyFunc: func(token *jwt.Token) (interface{}, error) {
			t, _, err := new(jwtv4.Parser).ParseUnverified(token.Raw, &SessionClaims{})
			if err != nil {
				return nil, err
			}
			return jwks.Keyfunc(t)
		},
	})
}

type SessionClaims struct {
	jwtv4.RegisteredClaims
	Session struct {
		Identity struct {
			Traits struct {
				DisplayName string `json:"display_name"`
				Email       string
			}
		}
	} `json:"session,omitempty"`
}

// RoleRepository provides role lookup by email (for config file)
type RoleRepository interface {
	GetRole(email string) string
}

// DatabaseRoleRepository provides role lookup by user ID (for database)
type DatabaseRoleRepository interface {
	GetUserRole(ctx context.Context, userID string) (string, error)
}

func Session(configRepo RoleRepository, dbRepo DatabaseRoleRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			sessionToken := &domain.SessionToken{
				Subject: "guest",
				Role:    domain.RoleGuest,
			}

			if ctx.Get("user") == nil {
				ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), domain.CtxSessionKey, sessionToken)))
				return next(ctx)
			}

			token := ctx.Get("user").(*jwt.Token)
			if claims, ok := token.Claims.(*SessionClaims); ok && token.Valid {
				sessionToken.Email = claims.Session.Identity.Traits.Email
				sessionToken.DisplayName = claims.Session.Identity.Traits.DisplayName
				sessionToken.Subject = claims.Subject
				sessionToken.CreatedAt = claims.IssuedAt.Time

				// First check config file (for dev/admin overrides)
				if configRepo != nil {
					role := configRepo.GetRole(sessionToken.Email)
					if role != "user" {
						sessionToken.Role = domain.Role(role)
					}
				}

				// Then check database if no special role from config
				if sessionToken.Role == "" || sessionToken.Role == domain.RoleUser || sessionToken.Role == domain.RoleGuest {
					if dbRepo != nil {
						dbRole, err := dbRepo.GetUserRole(ctx.Request().Context(), sessionToken.Subject)
						if err == nil && dbRole != "user" {
							sessionToken.Role = domain.Role(dbRole)
						}
					}
				}

				// Default to user if no special role found
				if sessionToken.Role == "" || sessionToken.Role == domain.RoleGuest {
					sessionToken.Role = domain.RoleUser
				}
			}

			ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), domain.CtxSessionKey, sessionToken)))

			// Allow banned users to check their role, but block everything else
			if sessionToken.Role == domain.RoleBanned && ctx.Path() != "/current-user/role" {
				return ctx.NoContent(http.StatusForbidden)
			}

			return next(ctx)
		}
	}
}
