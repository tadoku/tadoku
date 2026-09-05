package http

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/common/middleware"
)

// Exercise the legacy authentication middleware across real HTTP connections.
// The facade must not replace a user token with a privileged service identity.
func TestFacadePreservesLegacyIdentityAndAudienceChecks(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
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
			facade := httptest.NewServer(newTestHandler(t, upstream.URL, 5*time.Second))
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
						require.NoError(t, err)
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
						require.NoError(t, err)
						request.Header.Set("Authorization", authorization)
						response, err := (&stdhttp.Client{Timeout: 5 * time.Second}).Do(request)
						require.NoError(t, err)
						body, err := io.ReadAll(response.Body)
						_ = response.Body.Close()
						require.NoError(t, err)
						assert.Equal(t, test.status, response.StatusCode)
						if index == 0 {
							directBody = body
							if test.status == stdhttp.StatusOK {
								assert.Contains(t, string(body), `"Subject":"`+test.subject+`"`)
							}
						} else {
							assert.Equal(t, string(directBody), string(body))
						}
					}
				})
			}
		})
	}
}

func TestFacadePreservesLegacyResponses(t *testing.T) {
	for _, status := range []int{200, 204, 302, 400, 401, 403, 404, 409, 422, 429, 500, 503} {
		t.Run(stdhttp.StatusText(status), func(t *testing.T) {
			upstream := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.Header().Set("Location", "/next?return=%2Fcontests")
				w.Header().Set("Retry-After", "30")
				w.Header().Set("Cache-Control", "private, no-store")
				w.Header().Add("Set-Cookie", "first=one; HttpOnly; Secure")
				w.Header().Add("Set-Cookie", "second=two; SameSite=Lax")
				w.WriteHeader(status)
				if status != stdhttp.StatusNoContent {
					_, _ = io.WriteString(w, `{"detail":"legacy response"}`)
				}
			}))
			t.Cleanup(upstream.Close)
			facade := httptest.NewServer(newTestHandler(t, upstream.URL, 5*time.Second))
			t.Cleanup(facade.Close)
			client := &stdhttp.Client{
				Timeout:       5 * time.Second,
				CheckRedirect: func(*stdhttp.Request, []*stdhttp.Request) error { return stdhttp.ErrUseLastResponse },
			}
			direct, err := client.Get(upstream.URL + "/response")
			require.NoError(t, err)
			defer direct.Body.Close()
			proxied, err := client.Get(facade.URL + "/content/response")
			require.NoError(t, err)
			defer proxied.Body.Close()
			assert.Equal(t, direct.StatusCode, proxied.StatusCode)
			for _, header := range []string{"Content-Type", "Location", "Retry-After", "Cache-Control", "Set-Cookie"} {
				assert.Equal(t, direct.Header.Values(header), proxied.Header.Values(header), header)
			}
			directBody, err := io.ReadAll(direct.Body)
			require.NoError(t, err)
			proxiedBody, err := io.ReadAll(proxied.Body)
			require.NoError(t, err)
			assert.Equal(t, directBody, proxiedBody)
		})
	}
}

func TestFacadeStreamsResponseBeforeUpstreamFinishes(t *testing.T) {
	finish := make(chan struct{})
	upstream := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: first\n\n")
		_ = stdhttp.NewResponseController(w).Flush()
		select {
		case <-finish:
			_, _ = io.WriteString(w, "data: last\n\n")
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(upstream.Close)
	facade := httptest.NewServer(newTestHandler(t, upstream.URL, 5*time.Second))
	t.Cleanup(facade.Close)
	response, err := (&stdhttp.Client{Timeout: 5 * time.Second}).Get(facade.URL + "/immersion/events")
	require.NoError(t, err)
	defer response.Body.Close()
	first := make([]byte, len("data: first\n\n"))
	_, err = io.ReadFull(response.Body, first)
	require.NoError(t, err, "first chunk must arrive while the upstream is still waiting")
	assert.Equal(t, "data: first\n\n", string(first))
	close(finish)
	last, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	assert.Equal(t, "data: last\n\n", string(last))
}

func BenchmarkFacade(b *testing.B) {
	upstream := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = io.WriteString(w, `{"id":"example","title":"A small API response"}`)
	}))
	b.Cleanup(upstream.Close)
	facade := httptest.NewServer(newTestHandler(b, upstream.URL, 5*time.Second))
	b.Cleanup(facade.Close)
	client := &stdhttp.Client{Timeout: 5 * time.Second}
	for _, route := range []struct{ name, url string }{
		{"direct", upstream.URL + "/pages/example"},
		{"proxy", facade.URL + "/content/pages/example"},
	} {
		b.Run(route.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				response, err := client.Get(route.url)
				require.NoError(b, err)
				_, err = io.Copy(io.Discard, response.Body)
				_ = response.Body.Close()
				require.NoError(b, err)
				require.Equal(b, stdhttp.StatusOK, response.StatusCode)
			}
		})
	}
}
