package http

import (
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/app"
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

		ctx, err := s.application.AuthenticateEmailReader(request.Context(), token)
		if errors.Is(err, app.ErrEmailTenantNotServed) {
			writeJSON(w, stdhttp.StatusMisdirectedRequest, map[string]string{"message": err.Error()})
			return
		}
		if err != nil {
			writeResponseError(w, request, err)
			return
		}

		if observed, ok := ctx.Value(observedTenantKey{}).(*requestTenant); ok {
			observed.key, _ = tenant.FromContext(ctx)
		}
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}
