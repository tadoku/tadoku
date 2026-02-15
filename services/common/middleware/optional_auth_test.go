package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
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

type mockRolesService struct {
	claims roles.Claims
	err    error
}

func (m *mockRolesService) ClaimsForSubject(_ context.Context, _ string) (roles.Claims, error) {
	return m.claims, m.err
}

func (m *mockRolesService) ClaimsForSubjects(_ context.Context, ids []string) (map[string]roles.Claims, error) {
	result := make(map[string]roles.Claims, len(ids))
	for _, id := range ids {
		result[id] = m.claims
	}
	return result, m.err
}

// jwksJSON builds a minimal JWKS response from an RSA public key.
func jwksJSON(pub *rsa.PublicKey) []byte {
	b64 := func(b *big.Int) string {
		return base64.RawURLEncoding.EncodeToString(b.Bytes())
	}
	keys := map[string]interface{}{
		"keys": []map[string]string{
			{
				"kty": "RSA",
				"kid": "test-key",
				"alg": "RS256",
				"use": "sig",
				"n":   b64(pub.N),
				"e":   b64(big.NewInt(int64(pub.E))),
			},
		},
	}
	data, _ := json.Marshal(keys)
	return data
}

func setupJWKSServer(t *testing.T) (*rsa.PrivateKey, *httptest.Server) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(jwksJSON(&privateKey.PublicKey))
	}))
	t.Cleanup(server.Close)

	return privateKey, server
}

func signToken(t *testing.T, privateKey *rsa.PrivateKey, claims jwtv4.Claims) string {
	t.Helper()

	token := jwtv4.NewWithClaims(jwtv4.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"

	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)
	return tokenString
}

func TestOptionalAdminAuth_NoJWT_Allowed(t *testing.T) {
	_, jwksServer := setupJWKSServer(t)

	rolesSvc := &mockRolesService{}
	mw := OptionalAdminAuth(jwksServer.URL, rolesSvc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	nextCalled := false
	err := mw(func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOptionalAdminAuth_InvalidJWT_Unauthorized(t *testing.T) {
	_, jwksServer := setupJWKSServer(t)

	rolesSvc := &mockRolesService{}
	mw := OptionalAdminAuth(jwksServer.URL, rolesSvc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestOptionalAdminAuth_ValidAdminJWT_Allowed(t *testing.T) {
	privateKey, jwksServer := setupJWKSServer(t)

	rolesSvc := &mockRolesService{
		claims: roles.Claims{
			Subject:       "user-123",
			Authenticated: true,
			Admin:         true,
		},
	}
	mw := OptionalAdminAuth(jwksServer.URL, rolesSvc)

	claims := &UnifiedClaims{
		RegisteredClaims: jwtv4.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenString := signToken(t, privateKey, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	nextCalled := false
	err := mw(func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOptionalAdminAuth_ValidNonAdminJWT_Forbidden(t *testing.T) {
	privateKey, jwksServer := setupJWKSServer(t)

	rolesSvc := &mockRolesService{
		claims: roles.Claims{
			Subject:       "user-456",
			Authenticated: true,
			Admin:         false,
		},
	}
	mw := OptionalAdminAuth(jwksServer.URL, rolesSvc)

	claims := &UnifiedClaims{
		RegisteredClaims: jwtv4.RegisteredClaims{
			Subject:   "user-456",
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenString := signToken(t, privateKey, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestOptionalAdminAuth_BannedAdminJWT_Forbidden(t *testing.T) {
	privateKey, jwksServer := setupJWKSServer(t)

	rolesSvc := &mockRolesService{
		claims: roles.Claims{
			Subject:       "user-789",
			Authenticated: true,
			Admin:         true,
			Banned:        true,
		},
	}
	mw := OptionalAdminAuth(jwksServer.URL, rolesSvc)

	claims := &UnifiedClaims{
		RegisteredClaims: jwtv4.RegisteredClaims{
			Subject:   "user-789",
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenString := signToken(t, privateKey, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestOptionalAdminAuth_ServiceToken_Allowed(t *testing.T) {
	privateKey, jwksServer := setupJWKSServer(t)

	rolesSvc := &mockRolesService{}
	mw := OptionalAdminAuth(jwksServer.URL, rolesSvc)

	claims := &UnifiedClaims{
		RegisteredClaims: jwtv4.RegisteredClaims{
			Subject:   "system:serviceaccount:default:immersion-api",
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(time.Hour)),
		},
		Type: "service",
	}
	tokenString := signToken(t, privateKey, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	nextCalled := false
	err := mw(func(c echo.Context) error {
		nextCalled = true
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.True(t, nextCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestOptionalAdminAuth_GuestSubject_Unauthorized(t *testing.T) {
	privateKey, jwksServer := setupJWKSServer(t)

	rolesSvc := &mockRolesService{}
	mw := OptionalAdminAuth(jwksServer.URL, rolesSvc)

	claims := &UnifiedClaims{
		RegisteredClaims: jwtv4.RegisteredClaims{
			Subject:   "guest",
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenString := signToken(t, privateKey, claims)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(ctx)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
