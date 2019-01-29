package infra_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/infra"
	"github.com/tadoku/api/interfaces/services"
)

func TestRouter_RestrictedRoute(t *testing.T) {
	handler := func(ctx services.Context) error {
		return ctx.String(200, "test")
	}
	secret := "foobar"
	routes := []services.Route{
		{Method: http.MethodGet, Path: "/unrestricted", HandlerFunc: handler, Restricted: false},
		{Method: http.MethodGet, Path: "/restricted", HandlerFunc: handler, Restricted: true},
	}
	e := infra.NewRouter("1337", secret, routes...)

	for _, tc := range []struct {
		path          string
		expStatusCode int
	}{
		{
			path: "/restricted",
			// @TODO: This should be 401 instead of 400
			expStatusCode: http.StatusBadRequest,
		},
		{
			path:          "/unrestricted",
			expStatusCode: http.StatusOK,
		},
	} {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		res := httptest.NewRecorder()
		e.ServeHTTP(res, req)

		assert.Equal(t, tc.expStatusCode, res.Code)
	}
}
