package legacy_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	transporthttp "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
	"io"
	"log/slog"
	"math/big"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/middleware"
)

// Exercise the legacy authentication middleware across real HTTP connections.
// The facade must not replace a user token with a privileged service identity.
func TestFacadePreservesLegacyIdentityAndAudienceChecks(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	jwks := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "kid": "facade-test", "alg": "RS256", "use": "sig",
			"n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	}))
	t.Cleanup(jwks.Close)
	verifyJWT := middleware.VerifyJWT(jwks.URL)

	for _, service := range []string{"authz", "content", "immersion", "profile"} {
		t.Run(service, func(t *testing.T) {
			legacy := echo.New()
			legacy.Use(verifyJWT, middleware.Identity(), middleware.RequireServiceAudience(service+"-api"))
			legacy.GET("/identity", func(ctx echo.Context) error {
				return ctx.JSON(stdhttp.StatusOK, commondomain.ParseIdentity(ctx.Request().Context()))
			})
			upstream := httptest.NewServer(legacy)
			t.Cleanup(upstream.Close)
			facade := httptest.NewServer(proxyHandler(t, upstream.URL))
			t.Cleanup(facade.Close)

			for _, test := range []struct {
				name, subject, identityType, audience string
				status                                int
			}{
				{name: "guest", subject: "guest", status: 200},
				{name: "user", subject: "reader", status: 200},
				{name: "service", subject: "system:serviceaccount:dev:caller", identityType: "service", audience: service + "-api", status: 200},
				{name: "wrong audience", subject: "system:serviceaccount:dev:caller", identityType: "service", audience: "different-api", status: 403},
				{name: "missing token", status: 400},
				{name: "invalid signature", subject: "reader", status: 401},
			} {
				t.Run(test.name, func(t *testing.T) {
					authorization := ""
					if test.subject != "" {
						token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
							"sub": test.subject, "type": test.identityType, "aud": []string{test.audience},
							"iat": 1700000000,
						})
						token.Header["kid"] = "facade-test"
						signed, err := token.SignedString(key)
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if test.name == "invalid signature" {
							parts := strings.Split(signed, ".")
							parts[2] = base64.RawURLEncoding.EncodeToString([]byte("invalid"))
							signed = strings.Join(parts, ".")
						}
						authorization = "Bearer " + signed
					}
					var directBody []byte
					for index, url := range []string{upstream.URL + "/identity", facade.URL + "/" + service + "/identity"} {
						request, err := stdhttp.NewRequest(stdhttp.MethodGet, url, nil)
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
						request.Header.Set("Authorization", authorization)
						response, err := (&stdhttp.Client{Timeout: 5 * time.Second}).Do(request)
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
						body, err := io.ReadAll(response.Body)
						_ = response.Body.Close()
						if err != nil {
							t.Fatalf("unexpected error: %v", err)
						}
						if response.StatusCode != test.status {
							t.Errorf("got %v, want %v", response.StatusCode, test.status)
						}
						if index == 0 {
							directBody = body
							if test.status == stdhttp.StatusOK {
								if !strings.Contains(string(body), `"Subject":"`+test.subject+`"`) {
									t.Errorf("missing %v in %v", `"Subject":"`+test.subject+`"`, string(body))
								}
							}
						} else {
							if string(body) != string(directBody) {
								t.Errorf("got %v, want %v", string(body), string(directBody))
							}
						}
					}
				})
			}
		})
	}
}

func proxyHandler(t *testing.T, target string) stdhttp.Handler {
	t.Helper()
	handler, err := transporthttp.NewProxyHandler(transporthttp.Upstreams{Authz: target, Content: target, Immersion: target, Profile: target}, stdhttp.DefaultTransport, 5*time.Second, prometheus.NewRegistry(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return handler
}
