package middleware

import (
	"context"
	"fmt"

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
			return context.Path() == "/ping"
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

type RoleRepository interface {
	GetRole(string) string
}

func Session(repository RoleRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			sessionToken := &domain.SessionToken{
				Subject: "guest",
				Role:    "guest",
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
				sessionToken.Role = domain.Role(repository.GetRole(sessionToken.Email))
			}

			ctx.SetRequest(ctx.Request().WithContext(context.WithValue(ctx.Request().Context(), domain.CtxSessionKey, sessionToken)))
			return next(ctx)
		}
	}
}
