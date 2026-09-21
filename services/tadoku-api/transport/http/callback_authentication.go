package http

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	stdhttp "net/http"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/callbackauth"
)

func NewCallbackAuthentication(expectedToken string) (func(stdhttp.Handler) stdhttp.Handler, error) {
	if expectedToken == "" {
		return nil, fmt.Errorf("callback authentication token is required")
	}
	expectedHash := sha256.Sum256([]byte(expectedToken))

	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			scheme, token, found := strings.Cut(r.Header.Get("Authorization"), " ")
			providedHash := sha256.Sum256([]byte(token))
			if !found || !strings.EqualFold(scheme, "Bearer") || token == "" ||
				subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) != 1 {
				w.WriteHeader(stdhttp.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(callbackauth.WithAuthentication(r.Context())))
		})
	}, nil
}
