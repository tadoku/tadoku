package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/common/authz/roles"
)

func TestRejectBannedUsers_FailOpenWhenAuthzUnavailable(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/some-path")

	ctx.SetRequest(req.WithContext(roles.WithClaims(req.Context(), roles.Claims{
		Subject:       "kratos-id",
		Authenticated: true,
		Err:           errors.New("keto down"),
	})))

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	}

	err := RejectBannedUsers()(next)(ctx)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRejectBannedUsers_BlockedWhenBanned(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/some-path")

	ctx.SetRequest(req.WithContext(roles.WithClaims(req.Context(), roles.Claims{
		Subject:       "kratos-id",
		Authenticated: true,
		Banned:        true,
	})))

	err := RejectBannedUsers()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
