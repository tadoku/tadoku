package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestVerifyJWT_SkipsPing(t *testing.T) {
	_, jwksServer := setupJWKSServer(t)
	mw := VerifyJWT(jwksServer.URL)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/ping")

	nextCalled := false
	err := mw(func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusNoContent)
	})(ctx)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestVerifyJWT_MissingCredentials_BadRequest(t *testing.T) {
	_, jwksServer := setupJWKSServer(t)
	mw := VerifyJWT(jwksServer.URL)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/some-path")

	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(ctx)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Equal(t, "missing or malformed jwt", httpErr.Message)
}

func TestVerifyJWT_ValidToken_SetsUser(t *testing.T) {
	privateKey, jwksServer := setupJWKSServer(t)
	mw := VerifyJWT(jwksServer.URL)

	claims := &UnifiedClaims{
		RegisteredClaims: jwtv4.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenString := signToken(t, privateKey, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/some-path")

	nextCalled := false
	err := mw(func(c echo.Context) error {
		nextCalled = true
		token, ok := c.Get("user").(*jwtv4.Token)
		require.True(t, ok)
		assert.True(t, token.Valid)
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
}
