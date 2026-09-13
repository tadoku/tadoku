package http

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
)

type Authenticator struct{ keys *keyfunc.JWKS }

// NewAuthenticator fetches the same gateway JWKS used by Content. The supplied
// client's timeout bounds startup. Keys are held for this process lifetime,
// matching the legacy middleware (no background refresh was configured there).
func NewAuthenticator(ctx context.Context, jwksURL string, client *stdhttp.Client) (*Authenticator, error) {
	if client == nil || client.Timeout <= 0 {
		return nil, fmt.Errorf("bounded JWKS client is required")
	}
	keys, err := keyfunc.Get(jwksURL, keyfunc.Options{Ctx: ctx, Client: client})
	if err != nil {
		return nil, fmt.Errorf("fetch signing keys: %w", err)
	}
	return &Authenticator{keys: keys}, nil
}

func (a *Authenticator) Close() { a.keys.EndBackground() }

type claims struct {
	jwt.RegisteredClaims
	Type string `json:"type,omitempty"`
}

func (a *Authenticator) authenticate(request *stdhttp.Request) (app.Principal, int) {
	status := stdhttp.StatusBadRequest
	for index, header := range request.Header.Values("Authorization") {
		if len(header) <= len("Bearer ") || !strings.EqualFold(header[:len("Bearer ")], "Bearer ") {
			continue
		}
		value := &claims{}
		token, err := jwt.ParseWithClaims(header[len("Bearer "):], value, a.keys.Keyfunc)
		if err == nil && token.Valid {
			// Legacy user identity construction requires iat and otherwise returns
			// a recovered 500. Preserve that response without recreating its panic.
			if value.Type != "service" && value.IssuedAt == nil {
				return app.Principal{}, stdhttp.StatusInternalServerError
			}
			return app.Principal{Subject: value.Subject, Service: value.Type == "service", Audience: []string(value.Audience)}, 0
		}
		status = stdhttp.StatusUnauthorized
		// Echo's extractor considers at most 20 Authorization header values.
		if index >= 19 {
			break
		}
	}
	return app.Principal{}, status
}
