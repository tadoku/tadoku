package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// NewJWTAuthentication accepts only signed guest or UUID subjects. Callers must
// enforce roles, bans, permissions and service audiences separately.
func NewJWTAuthentication(lifetime context.Context, jwksURL string, timeout, maxTokenAge time.Duration, issuer string, logger *slog.Logger) (func(stdhttp.Handler) stdhttp.Handler, error) {
	if lifetime == nil {
		return nil, fmt.Errorf("authentication lifetime context is required")
	}
	if jwksURL == "" {
		return nil, fmt.Errorf("JWKS URL is required")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("JWKS fetch timeout must be positive")
	}
	if maxTokenAge <= 0 {
		return nil, fmt.Errorf("maximum token age must be positive")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	keys, err := keyfunc.Get(jwksURL, keyfunc.Options{
		Ctx:              lifetime,
		RefreshTimeout:   timeout,
		RefreshInterval:  time.Hour,
		RefreshRateLimit: time.Minute,
		RefreshErrorHandler: func(err error) {
			logger.Error("jwks refresh failed", "error", err)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("fetch authentication JWKS: %w", err)
	}
	context.AfterFunc(lifetime, keys.EndBackground)

	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			keyForToken := func(token *jwt.Token) (interface{}, error) {
				key, err := keys.Keyfunc(token)
				if !errors.Is(err, keyfunc.ErrKIDNotFound) {
					return key, err
				}

				refreshContext, cancelRefresh := context.WithCancel(r.Context())
				stopLifetimeCancellation := context.AfterFunc(lifetime, cancelRefresh)
				defer stopLifetimeCancellation()
				defer cancelRefresh()
				if err := keys.Refresh(refreshContext, keyfunc.RefreshOptions{}); err != nil {
					return nil, err
				}
				return keys.Keyfunc(token)
			}

			foundToken := false
			for i, header := range r.Header.Values("Authorization") {
				if len(header) <= len("Bearer ") || !strings.EqualFold(header[:len("Bearer ")], "Bearer ") {
					continue
				}
				foundToken = true

				claims := &userClaims{}
				token, err := jwt.ParseWithClaims(header[len("Bearer "):], claims, keyForToken, jwt.WithValidMethods([]string{"RS256"}))
				if err == nil &&
					token.Valid &&
					claims.ExpiresAt != nil &&
					claims.IssuedAt != nil &&
					!claims.IssuedAt.Time.Add(maxTokenAge).Before(jwt.TimeFunc()) &&
					(issuer == "" || claims.Issuer == issuer) &&
					claims.Type != "service" &&
					(claims.Subject == "guest" || uuid.Validate(claims.Subject) == nil) {
					user := &identity.User{
						Subject:     claims.Subject,
						DisplayName: claims.Session.Identity.Traits.DisplayName,
						Email:       claims.Session.Identity.Traits.Email,
						CreatedAt:   claims.IssuedAt.Time,
					}
					next.ServeHTTP(w, r.WithContext(identity.WithUser(r.Context(), user)))
					return
				}

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
