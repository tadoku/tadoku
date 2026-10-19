package http

import (
	stdhttp "net/http"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func emailAccessToken(request *stdhttp.Request) (string, bool) {
	scheme, value, ok := strings.Cut(request.Header.Get("Authorization"), " ")
	return value, ok && strings.EqualFold(scheme, "Bearer") && strings.HasPrefix(value, "tdk_at_")
}

func (s *server) authenticateEmailToken(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, request *stdhttp.Request) {
		token, ok := emailAccessToken(request)
		if !ok {
			w.WriteHeader(stdhttp.StatusUnauthorized)
			return
		}

		key, err := s.application.AuthenticateEmailReader(request.Context(), token)
		if err != nil {
			writeResponseError(w, request, err)
			return
		}

		next.ServeHTTP(w, request.WithContext(tenant.WithKey(request.Context(), key)))
	})
}
