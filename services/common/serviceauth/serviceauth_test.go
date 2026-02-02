package serviceauth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestKeyPair generates an ECDSA P-256 key pair for testing
func generateTestKeyPair(t *testing.T) (*ecdsa.PrivateKey, []byte, []byte) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Encode private key to PEM
	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privBytes,
	})

	// Encode public key to PEM
	pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return privateKey, privPEM, pubPEM
}

func TestNewTokenGenerator(t *testing.T) {
	_, privPEM, _ := generateTestKeyPair(t)

	t.Run("valid key", func(t *testing.T) {
		gen, err := NewTokenGenerator("test-service", privPEM)
		require.NoError(t, err)
		assert.Equal(t, "test-service", gen.ServiceName())
	})

	t.Run("invalid PEM", func(t *testing.T) {
		_, err := NewTokenGenerator("test-service", []byte("not a valid key"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse private key")
	})

	t.Run("empty key", func(t *testing.T) {
		_, err := NewTokenGenerator("test-service", []byte{})
		assert.Error(t, err)
	})
}

func TestTokenGeneratorGenerate(t *testing.T) {
	_, privPEM, _ := generateTestKeyPair(t)

	gen, err := NewTokenGenerator("issuer-service", privPEM)
	require.NoError(t, err)

	token, err := gen.Generate("target-service")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should have three parts (header.payload.signature)
	parts := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts++
		}
	}
	assert.Equal(t, 2, parts, "JWT should have 3 parts separated by 2 dots")
}

func TestTokenValidation(t *testing.T) {
	privateKey, privPEM, _ := generateTestKeyPair(t)

	gen, err := NewTokenGenerator("caller-service", privPEM)
	require.NoError(t, err)

	publicKeys := map[string]*ecdsa.PublicKey{
		"caller-service": &privateKey.PublicKey,
	}
	validator := NewTokenValidatorWithKeys("receiver-service", publicKeys)

	t.Run("valid token", func(t *testing.T) {
		token, err := gen.Generate("receiver-service")
		require.NoError(t, err)

		caller, err := validator.Validate(token)
		require.NoError(t, err)
		assert.Equal(t, "caller-service", caller)
	})

	t.Run("wrong audience", func(t *testing.T) {
		token, err := gen.Generate("other-service")
		require.NoError(t, err)

		_, err = validator.Validate(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not intended for this service")
	})

	t.Run("unknown issuer", func(t *testing.T) {
		// Create a token from an unknown service
		_, unknownPrivPEM, _ := generateTestKeyPair(t)
		unknownGen, err := NewTokenGenerator("unknown-service", unknownPrivPEM)
		require.NoError(t, err)

		token, err := unknownGen.Generate("receiver-service")
		require.NoError(t, err)

		_, err = validator.Validate(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown issuer")
	})

	t.Run("invalid signature", func(t *testing.T) {
		// Create a token signed with a different key
		_, otherPrivPEM, _ := generateTestKeyPair(t)
		otherGen, err := NewTokenGenerator("caller-service", otherPrivPEM)
		require.NoError(t, err)

		token, err := otherGen.Generate("receiver-service")
		require.NoError(t, err)

		// This should fail because the signature doesn't match the registered public key
		_, err = validator.Validate(token)
		assert.Error(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		// Create a generator with a custom clock that returns a time in the past
		expiredGen, err := NewTokenGenerator("caller-service", privPEM)
		require.NoError(t, err)
		expiredGen.clock = func() time.Time {
			return time.Now().Add(-10 * time.Minute)
		}

		token, err := expiredGen.Generate("receiver-service")
		require.NoError(t, err)

		_, err = validator.Validate(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := validator.Validate("not.a.valid.token")
		assert.Error(t, err)
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := validator.Validate("")
		assert.Error(t, err)
	})
}

func TestServiceAuthMiddleware(t *testing.T) {
	privateKey, privPEM, _ := generateTestKeyPair(t)

	gen, err := NewTokenGenerator("caller-service", privPEM)
	require.NoError(t, err)

	publicKeys := map[string]*ecdsa.PublicKey{
		"caller-service": &privateKey.PublicKey,
	}
	validator := NewTokenValidatorWithKeys("receiver-service", publicKeys)

	handler := func(c echo.Context) error {
		caller := GetCallingService(c)
		return c.String(http.StatusOK, "caller: "+caller)
	}

	t.Run("valid service token", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		token, err := gen.Generate("receiver-service")
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		c := e.NewContext(req, rec)
		middleware := ServiceAuth(validator)
		err = middleware(handler)(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "caller: caller-service", rec.Body.String())
	})

	t.Run("missing authorization header", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		middleware := ServiceAuth(validator)
		err := middleware(handler)(c)

		require.NoError(t, err) // Handler returns JSON response, not error
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		req.Header.Set("Authorization", "Bearer invalid-token")

		c := e.NewContext(req, rec)
		middleware := ServiceAuth(validator)
		err := middleware(handler)(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestGetCallingService(t *testing.T) {
	t.Run("no calling service in context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		caller := GetCallingService(c)
		assert.Empty(t, caller)
	})
}

func TestIsServiceRequest(t *testing.T) {
	t.Run("not a service request", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.False(t, IsServiceRequest(c))
	})
}

func TestClaims(t *testing.T) {
	now := time.Now()
	claims := NewServiceClaims("issuer", "audience", now)

	assert.Equal(t, "issuer", claims.Issuer)
	assert.Equal(t, "issuer", claims.Subject, "subject should match issuer per RFC 7523")
	assert.Equal(t, []string{"audience"}, []string(claims.Audience))
	assert.Equal(t, now.Unix(), claims.IssuedAt.Unix())
	assert.Equal(t, now.Unix(), claims.NotBefore.Unix())
	assert.Equal(t, now.Add(TokenExpiry).Unix(), claims.ExpiresAt.Unix())
}

func TestInputValidation(t *testing.T) {
	_, privPEM, _ := generateTestKeyPair(t)

	t.Run("empty service name in generator", func(t *testing.T) {
		_, err := NewTokenGenerator("", privPEM)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service name cannot be empty")
	})

	t.Run("empty target service in generate", func(t *testing.T) {
		gen, err := NewTokenGenerator("test-service", privPEM)
		require.NoError(t, err)

		_, err = gen.Generate("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target service cannot be empty")
	})
}

func TestWithExpiry(t *testing.T) {
	privateKey, privPEM, _ := generateTestKeyPair(t)

	t.Run("custom expiry", func(t *testing.T) {
		gen, err := NewTokenGenerator("caller", privPEM)
		require.NoError(t, err)
		gen.WithExpiry(1 * time.Minute)

		// Generate token with custom expiry
		token, err := gen.Generate("receiver")
		require.NoError(t, err)

		// Validate the token
		publicKeys := map[string]*ecdsa.PublicKey{
			"caller": &privateKey.PublicKey,
		}
		validator := NewTokenValidatorWithKeys("receiver", publicKeys)

		caller, err := validator.Validate(token)
		require.NoError(t, err)
		assert.Equal(t, "caller", caller)
	})
}

func TestNewTokenGeneratorFromFile(t *testing.T) {
	_, privPEM, _ := generateTestKeyPair(t)

	t.Run("valid file", func(t *testing.T) {
		// Create temp file with private key
		tmpFile, err := os.CreateTemp("", "test-key-*.pem")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(privPEM)
		require.NoError(t, err)
		tmpFile.Close()

		gen, err := NewTokenGeneratorFromFile("test-service", tmpFile.Name())
		require.NoError(t, err)
		assert.Equal(t, "test-service", gen.ServiceName())
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := NewTokenGeneratorFromFile("test-service", "/nonexistent/path/key.pem")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read private key file")
	})
}

func TestNewTokenValidator(t *testing.T) {
	_, _, pubPEM := generateTestKeyPair(t)

	t.Run("valid directory with keys", func(t *testing.T) {
		// Create temp directory with public key
		tmpDir, err := os.MkdirTemp("", "test-keys-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		keyPath := tmpDir + "/caller-service.pub"
		err = os.WriteFile(keyPath, pubPEM, 0644)
		require.NoError(t, err)

		validator, err := NewTokenValidator("receiver-service", tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, validator)
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-keys-empty-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		_, err = NewTokenValidator("receiver-service", tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no public keys found")
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := NewTokenValidator("receiver-service", "/nonexistent/path")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read public key directory")
	})

	t.Run("empty service name", func(t *testing.T) {
		_, err := NewTokenValidator("", "/some/path")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service name cannot be empty")
	})
}

func TestServiceOrUserAuthMiddleware(t *testing.T) {
	privateKey, privPEM, _ := generateTestKeyPair(t)

	gen, err := NewTokenGenerator("caller-service", privPEM)
	require.NoError(t, err)

	publicKeys := map[string]*ecdsa.PublicKey{
		"caller-service": &privateKey.PublicKey,
	}
	validator := NewTokenValidatorWithKeys("receiver-service", publicKeys)

	handler := func(c echo.Context) error {
		caller := GetCallingService(c)
		if caller != "" {
			return c.String(http.StatusOK, "service: "+caller)
		}
		return c.String(http.StatusOK, "user request")
	}

	t.Run("valid service token", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		token, err := gen.Generate("receiver-service")
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		c := e.NewContext(req, rec)
		middleware := ServiceOrUserAuth(validator, nil, nil)
		err = middleware(handler)(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "service: caller-service", rec.Body.String())
	})

	t.Run("invalid bearer token rejected", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		req.Header.Set("Authorization", "Bearer invalid-token")

		c := e.NewContext(req, rec)
		middleware := ServiceOrUserAuth(validator, nil, nil)
		err := middleware(handler)(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("no bearer token falls through to session", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		middleware := ServiceOrUserAuth(validator, nil, nil)
		err := middleware(handler)(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "user request", rec.Body.String())
	})
}
