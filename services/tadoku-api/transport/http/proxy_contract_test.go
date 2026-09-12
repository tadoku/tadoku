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
	"slices"
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
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer direct.Body.Close()
			proxied, err := client.Get(facade.URL + "/content/response")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer proxied.Body.Close()
			if proxied.StatusCode != direct.StatusCode {
				t.Errorf("got %v, want %v", proxied.StatusCode, direct.StatusCode)
			}
			for _, header := range []string{"Content-Type", "Location", "Retry-After", "Cache-Control", "Set-Cookie"} {
				if !slices.Equal(direct.Header.Values(header), proxied.Header.Values(header)) {
					t.Errorf("got %v, want %v", proxied.Header.Values(header), direct.Header.Values(header))
				}
			}
			directBody, err := io.ReadAll(direct.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			proxiedBody, err := io.ReadAll(proxied.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(directBody, proxiedBody) {
				t.Errorf("got %v, want %v", proxiedBody, directBody)
			}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer response.Body.Close()
	first := make([]byte, len("data: first\n\n"))
	_, err = io.ReadFull(response.Body, first)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(first) != "data: first\n\n" {
		t.Errorf("got %v, want %v", string(first), "data: first\n\n")
	}
	close(finish)
	last, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(last) != "data: last\n\n" {
		t.Errorf("got %v, want %v", string(last), "data: last\n\n")
	}
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
				if err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
				_, err = io.Copy(io.Discard, response.Body)
				_ = response.Body.Close()
				if err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
				if response.StatusCode != stdhttp.StatusOK {
					b.Fatalf("got %v, want %v", response.StatusCode, stdhttp.StatusOK)
				}
			}
		})
	}
}
