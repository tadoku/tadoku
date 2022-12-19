package middleware

import (
	"fmt"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
		KeyFunc: func(token *jwt.Token) (interface{}, error) {
			t, _, err := new(jwtv4.Parser).ParseUnverified(token.Raw, jwtv4.MapClaims{})
			if err != nil {
				return nil, err
			}
			return jwks.Keyfunc(t)
		},
	})
}

type SessionToken struct {
	Subject     string
	DisplayName string
	Email       string
	Role        string
}

func Session() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if ctx.Get("user") == nil {
				return next(ctx)
			}

			token := ctx.Get("user").(*jwt.Token)
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				sessionToken := &SessionToken{
					Subject: "guest",
					Role:    "guest",
				}
				if subject, ok := claims["sub"]; ok {
					if val, ok := subject.(string); ok {
						sessionToken.Subject = val
					}
				}
				if session, ok := claims["session"]; ok {
					if identity, ok := session.(map[string]interface{})["identity"]; ok {
						if traits, ok := identity.(map[string]interface{})["traits"]; ok {
							t := traits.(map[string]interface{})
							if displayName, ok := t["display_name"]; ok {
								if val, ok := displayName.(string); ok {
									sessionToken.DisplayName = val
								}
							}
							if email, ok := t["email"]; ok {
								if val, ok := email.(string); ok {
									sessionToken.Email = val
								}
							}
						}
					}
				}

				ctx.Set("session", sessionToken)
			}

			return next(ctx)
		}
	}
}
