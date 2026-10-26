package http

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app/api"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestAuthenticationRequiresConfiguration(t *testing.T) {
	for _, test := range []struct {
		name      string
		lifetime  context.Context
		url       string
		timeout   time.Duration
		maxAge    time.Duration
		logger    *slog.Logger
		wantError string
	}{
		{name: "missing lifetime", url: "http://jwks.test", timeout: time.Second, maxAge: 24 * time.Hour, logger: slog.Default(), wantError: "authentication lifetime context is required"},
		{name: "missing URL", lifetime: t.Context(), timeout: time.Second, maxAge: 24 * time.Hour, logger: slog.Default(), wantError: "JWKS URL is required"},
		{name: "missing timeout", lifetime: t.Context(), url: "http://jwks.test", maxAge: 24 * time.Hour, logger: slog.Default(), wantError: "JWKS fetch timeout must be positive"},
		{name: "negative timeout", lifetime: t.Context(), url: "http://jwks.test", timeout: -time.Second, maxAge: 24 * time.Hour, logger: slog.Default(), wantError: "JWKS fetch timeout must be positive"},
		{name: "missing maximum token age", lifetime: t.Context(), url: "http://jwks.test", timeout: time.Second, logger: slog.Default(), wantError: "maximum token age must be positive"},
		{name: "negative maximum token age", lifetime: t.Context(), url: "http://jwks.test", timeout: time.Second, maxAge: -time.Hour, logger: slog.Default(), wantError: "maximum token age must be positive"},
		{name: "missing logger", lifetime: t.Context(), url: "http://jwks.test", timeout: time.Second, maxAge: 24 * time.Hour, wantError: "logger is required"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewJWTAuthentication(test.lifetime, test.url, test.timeout, test.maxAge, "", tenant.Deployment{}, test.logger)
			if err == nil || err.Error() != test.wantError {
				t.Errorf("error=%v, want %q", err, test.wantError)
			}
		})
	}

	passthrough := func(next stdhttp.Handler) stdhttp.Handler { return next }
	_, err := NewHandler(api.New(api.Dependencies{}), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), nil, passthrough, passthrough)
	if err == nil {
		t.Error("router accepted missing authentication middleware")
	}

	_, err = NewHandler(api.New(api.Dependencies{}), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), passthrough, nil, passthrough)
	if err == nil {
		t.Error("router accepted missing banned-user middleware")
	}

	_, err = NewHandler(api.New(api.Dependencies{}), func(context.Context) error { return nil }, time.Second, prometheus.NewRegistry(), slog.Default(), passthrough, passthrough, nil)
	if err == nil {
		t.Error("router accepted missing callback authentication middleware")
	}
}

func TestAuthenticationTokenPolicy(t *testing.T) {
	oldTimeFunc := jwt.TimeFunc
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	jwt.TimeFunc = func() time.Time { return now }
	defer func() { jwt.TimeFunc = oldTimeFunc }()

	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	edPublicKey, edPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	jwks := fmt.Sprintf(`{"keys":[%s,{"kty":"OKP","crv":"Ed25519","kid":"ed-key","alg":"EdDSA","use":"sig","x":%q}]}`,
		rsaJWK("rsa-key", &rsaPrivateKey.PublicKey), base64.RawURLEncoding.EncodeToString(edPublicKey))
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = w.Write([]byte(jwks))
	}))
	t.Cleanup(server.Close)

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", tenant.Deployment{}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	handler := authenticate(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	}))

	userID := "22222222-2222-4222-8222-222222222222"
	for _, test := range []struct {
		name       string
		issuedAt   time.Time
		issuer     string
		subject    string
		method     jwt.SigningMethod
		kid        string
		privateKey any
		want       int
	}{
		{name: "exact maximum age", issuedAt: now.Add(-24 * time.Hour), subject: userID, method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusNoContent},
		{name: "just beyond maximum age", issuedAt: now.Add(-24*time.Hour - time.Second), subject: userID, method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusUnauthorized},
		{name: "other issuer when unpinned", issuedAt: now.Add(-time.Minute), issuer: "https://other.example.test/", subject: userID, method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusNoContent},
		{name: "missing issuer when unpinned", issuedAt: now.Add(-time.Minute), subject: userID, method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusNoContent},
		{name: "non RS256 with matching key", issuedAt: now.Add(-time.Minute), subject: userID, method: jwt.SigningMethodEdDSA, kid: "ed-key", privateKey: edPrivateKey, want: stdhttp.StatusUnauthorized},
		{name: "guest subject", issuedAt: now.Add(-time.Minute), subject: "guest", method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusNoContent},
		{name: "non UUID subject", issuedAt: now.Add(-time.Minute), subject: "policy-user", method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusUnauthorized},
		{name: "missing subject", issuedAt: now.Add(-time.Minute), method: jwt.SigningMethodRS256, kid: "rsa-key", privateKey: rsaPrivateKey, want: stdhttp.StatusUnauthorized},
	} {
		t.Run(test.name, func(t *testing.T) {
			token := jwt.NewWithClaims(test.method, &userClaims{
				Tenant: "tadoku/prod",
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    test.issuer,
					Subject:   test.subject,
					IssuedAt:  jwt.NewNumericDate(test.issuedAt),
					ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				},
			})
			token.Header["kid"] = test.kid
			signed, err := token.SignedString(test.privateKey)
			if err != nil {
				t.Fatal(err)
			}

			request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
			request.Header.Set("Authorization", "Bearer "+signed)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestNewApplicationRoutesInheritSharedMiddleware(t *testing.T) {
	type authenticationContextKey struct{}

	var authenticationCompositions, banCompositions, authenticated, checkedBans int
	authenticatedRequest := authenticationContextKey{}
	authenticate := func(next stdhttp.Handler) stdhttp.Handler {
		authenticationCompositions++
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			authenticated++
			if _, ok := r.Context().Deadline(); !ok {
				t.Error("authentication did not inherit the request timeout")
			}
			if r.URL.Path == "/test/auth-failure" {
				w.WriteHeader(stdhttp.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authenticatedRequest, true)))
		})
	}
	rejectBanned := func(next stdhttp.Handler) stdhttp.Handler {
		banCompositions++
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			checkedBans++
			if authenticated, _ := r.Context().Value(authenticatedRequest).(bool); !authenticated {
				t.Error("ban check ran before authentication")
			}
			if r.URL.Path != "/test/a/b" {
				w.WriteHeader(stdhttp.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	router, err := NewHandler(
		api.New(api.Dependencies{}),
		func(context.Context) error { return nil },
		time.Second,
		prometheus.NewRegistry(),
		slog.Default(),
		authenticate,
		rejectBanned,
		func(next stdhttp.Handler) stdhttp.Handler { return next },
	)
	if err != nil {
		t.Fatal(err)
	}
	if authenticationCompositions != 1 || banCompositions != 1 {
		t.Fatalf("authentication compositions=%d ban compositions=%d, want 1 each", authenticationCompositions, banCompositions)
	}
	for _, path := range []string{"/test/first", "/test/second"} {
		router.Handle("GET "+path, stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
			t.Error("application handler bypassed the shared ban check")
			w.WriteHeader(stdhttp.StatusNoContent)
		}))
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, path, nil))
		if response.Code != stdhttp.StatusForbidden {
			t.Errorf("%s status=%d, want 403", path, response.Code)
		}
	}
	router.HandleFunc("GET /test/auth-failure", func(stdhttp.ResponseWriter, *stdhttp.Request) {
		t.Error("application handler bypassed failed authentication")
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/test/auth-failure", nil))
	if response.Code != stdhttp.StatusUnauthorized {
		t.Errorf("auth failure status=%d, want 401", response.Code)
	}
	router.HandleFunc("GET /test/{value}", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if value := r.PathValue("value"); value != "a/b" {
			t.Errorf("path value=%q, want %q", value, "a/b")
		}
		w.WriteHeader(stdhttp.StatusNoContent)
	})
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/test/a%2Fb", nil))
	if response.Code != stdhttp.StatusNoContent {
		t.Errorf("escaped path status=%d, want 204", response.Code)
	}
	if authenticationCompositions != 1 || banCompositions != 1 {
		t.Errorf("authentication compositions=%d ban compositions=%d, want 1 each", authenticationCompositions, banCompositions)
	}
	if authenticated != 4 || checkedBans != 3 {
		t.Errorf("authentication calls=%d ban checks=%d, want 4 and 3", authenticated, checkedBans)
	}
}

func TestCallbackAuthenticationIsScopedToExactOperation(t *testing.T) {
	callback, err := NewCallbackAuthentication("callback-token")
	if err != nil {
		t.Fatal(err)
	}
	businessAuthentication := func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
			w.WriteHeader(stdhttp.StatusUnauthorized)
		})
	}
	passthrough := func(next stdhttp.Handler) stdhttp.Handler { return next }
	router, err := NewHandler(
		api.New(api.Dependencies{}),
		func(context.Context) error { return nil },
		time.Second,
		prometheus.NewRegistry(),
		slog.Default(),
		businessAuthentication,
		passthrough,
		callback,
	)
	if err != nil {
		t.Fatal(err)
	}
	router.HandleFunc("PUT /authz/internal/v1/proxy/admin-check", func(stdhttp.ResponseWriter, *stdhttp.Request) {
		t.Error("callback credential bypassed business authentication for another method")
	})
	router.HandleFunc("POST /authz/internal/v1/proxy/admin-check/{extra}", func(stdhttp.ResponseWriter, *stdhttp.Request) {
		t.Error("callback credential bypassed business authentication for a subpath")
	})

	for _, target := range []struct {
		method string
		path   string
	}{
		{method: stdhttp.MethodPut, path: "/authz/internal/v1/proxy/admin-check"},
		{method: stdhttp.MethodPost, path: "/authz/internal/v1/proxy/admin-check/extra"},
	} {
		request := httptest.NewRequest(target.method, target.path, nil)
		request.Header.Set("Authorization", "Bearer callback-token")
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != stdhttp.StatusUnauthorized {
			t.Errorf("%s %s status=%d, want %d", target.method, target.path, response.Code, stdhttp.StatusUnauthorized)
		}
	}
}

func TestAuthenticationTenantClaims(t *testing.T) {
	now := time.Now()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = fmt.Fprintf(w, `{"keys":[%s]}`, rsaJWK("tenant-key", &privateKey.PublicKey))
	}))
	t.Cleanup(server.Close)

	authenticate, err := NewJWTAuthentication(t.Context(), server.URL, time.Second, 24*time.Hour, "", tenant.Deployment{}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	handler := authenticate(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		key, ok := tenant.FromContext(r.Context())
		if !ok {
			t.Error("verified handler has no tenant")
		}
		w.Header().Set("X-Test-Tenant", key.String())
		w.WriteHeader(stdhttp.StatusNoContent)
	}))

	for _, test := range []struct {
		name    string
		value   any
		present bool
		want    int
	}{
		{name: "missing", want: stdhttp.StatusUnauthorized},
		{name: "empty", present: true, value: "", want: stdhttp.StatusUnauthorized},
		{name: "bare", present: true, value: "tadoku", want: stdhttp.StatusUnauthorized},
		{name: "uppercase", present: true, value: "TADOKU/prod", want: stdhttp.StatusUnauthorized},
		{name: "extra slash", present: true, value: "e2e/run/id", want: stdhttp.StatusUnauthorized},
		{name: "non-string", present: true, value: []string{"tadoku/prod"}, want: stdhttp.StatusUnauthorized},
		{name: "production", present: true, value: "tadoku/prod", want: stdhttp.StatusNoContent},
		{name: "test", present: true, value: "e2e/tenant-claim", want: stdhttp.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			claims := jwt.MapClaims{"sub": "guest", "iat": now.Unix(), "exp": now.Add(time.Hour).Unix()}
			if test.present {
				claims["tenant"] = test.value
			}

			token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
			token.Header["kid"] = "tenant-key"
			signed, err := token.SignedString(privateKey)
			if err != nil {
				t.Fatal(err)
			}

			request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
			request.Header.Set("Authorization", "Bearer "+signed)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if test.want == stdhttp.StatusNoContent && response.Header().Get("X-Test-Tenant") != test.value {
				t.Errorf("context tenant=%q, want %q", response.Header().Get("X-Test-Tenant"), test.value)
			}
			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
			if test.want == stdhttp.StatusUnauthorized && response.Body.String() != "{\"message\":\"invalid or expired jwt\"}\n" {
				t.Errorf("body=%q", response.Body.String())
			}
		})
	}
}
