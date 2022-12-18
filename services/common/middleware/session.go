package middleware

import (
	"fmt"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func Session(jwksURL string) echo.MiddlewareFunc {
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
