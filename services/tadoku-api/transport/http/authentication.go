package http

import (
	"fmt"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// NewAuthentication loads the gateway's signing keys once and verifies user JWTs.
// It performs no role, ban, permission, or service-audience checks.
func NewAuthentication(jwksURL string, timeout time.Duration) (func(stdhttp.Handler) stdhttp.Handler, error) {
	if jwksURL == "" || timeout <= 0 {
		return nil, fmt.Errorf("JWKS URL and positive fetch timeout are required")
	}

	keys, err := keyfunc.Get(jwksURL, keyfunc.Options{RefreshTimeout: timeout})
	if err != nil {
		return nil, fmt.Errorf("fetch authentication JWKS: %w", err)
	}

	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			foundToken := false
			for i, header := range r.Header.Values("Authorization") {
				if len(header) <= len("Bearer ") || !strings.EqualFold(header[:len("Bearer ")], "Bearer ") {
					continue
				}
				foundToken = true

				claims := &userClaims{}
				token, err := jwt.ParseWithClaims(header[len("Bearer "):], claims, keys.Keyfunc)
				if err == nil && token.Valid && claims.IssuedAt != nil && claims.Type != "service" {
					user := &identity.User{
						Subject:     claims.Subject,
						DisplayName: claims.Session.Identity.Traits.DisplayName,
						Email:       claims.Session.Identity.Traits.Email,
						CreatedAt:   claims.IssuedAt.Time,
					}
					next.ServeHTTP(w, r.WithContext(identity.WithUser(r.Context(), user)))
					return
				}

				// Match Echo's header extractor: stop at the first bearer value at
				// index 19 or later. Non-bearer values are skipped before this limit.
				if i >= 19 {
					break
				}
			}

			status, message := stdhttp.StatusBadRequest, "missing or malformed jwt"
			if foundToken {
				status, message = stdhttp.StatusUnauthorized, "invalid or expired jwt"
			}
			writeJSON(w, status, struct {
				Message string `json:"message"`
			}{Message: message})
		})
	}, nil
}

type userClaims struct {
	jwt.RegisteredClaims
	Type    string `json:"type,omitempty"`
	Session struct {
		Identity struct {
			Traits struct {
				DisplayName string `json:"display_name"`
				Email       string `json:"email"`
			} `json:"traits"`
		} `json:"identity"`
	} `json:"session"`
}
