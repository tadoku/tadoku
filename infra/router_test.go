package infra_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/infra"
	"github.com/tadoku/api/interfaces/services"
	"github.com/tadoku/api/usecases"
)

func TestRouter_RestrictedRoute(t *testing.T) {
	handler := func(ctx services.Context) error {
		return ctx.String(200, "test")
	}
	secret := "foobar"
	routes := []services.Route{
		{Method: http.MethodGet, Path: "/unrestricted", HandlerFunc: handler},
		{Method: http.MethodGet, Path: "/restricted", HandlerFunc: handler, MinRole: domain.RoleUser},
		{Method: http.MethodGet, Path: "/registered_only", HandlerFunc: handler, MinRole: domain.RoleUser},
		{Method: http.MethodGet, Path: "/admin", HandlerFunc: handler, MinRole: domain.RoleAdmin},
	}
	e := infra.NewRouter("1337", secret, nil, nil, routes...)
	clock, _ := infra.NewClock("UTC")
	gen := infra.NewJWTGenerator(secret, clock)

	for _, tc := range []struct {
		path          string
		expStatusCode int
		user          *domain.User
		info          string
	}{
		{
			path:          "/restricted",
			expStatusCode: http.StatusUnauthorized,
			info:          "Missing JWT",
		},
		{
			path:          "/unrestricted",
			expStatusCode: http.StatusOK,
			info:          "Access to unrestricted page without token",
		},
		{
			path:          "/admin",
			expStatusCode: http.StatusForbidden,
			user:          &domain.User{Role: domain.RoleUser},
			info:          "No admin access as user",
		},
		{
			path:          "/admin",
			expStatusCode: http.StatusOK,
			user:          &domain.User{Role: domain.RoleAdmin},
			info:          "Admin access as admin",
		},
	} {
		token, _ := gen.NewToken(time.Hour*1, usecases.SessionClaims{User: tc.user})
		authHeader := middleware.DefaultJWTConfig.AuthScheme + " " + token

		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		req.Header.Set(echo.HeaderAuthorization, authHeader)

		res := httptest.NewRecorder()

		e.ServeHTTP(res, req)

		assert.Equal(t, tc.expStatusCode, res.Code, tc.info)
	}
}
