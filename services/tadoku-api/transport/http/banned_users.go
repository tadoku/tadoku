package http

import (
	"context"
	"log/slog"
	stdhttp "net/http"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

// RejectBannedUsers blocks authenticated users with the app:tadoku#banned relation.
// Provider failures deliberately fail open to preserve the existing API policy.
func RejectBannedUsers(
	check func(context.Context, string) (bool, error),
	logger *slog.Logger,
) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			user := identity.FromContext(r.Context())
			if user == nil || user.Subject == "" || user.Subject == "guest" {
				next.ServeHTTP(w, r)
				return
			}

			banned, err := check(r.Context(), user.Subject)
			if err != nil {
				logger.Error("banned-user check unavailable; allowing request", "subject", user.Subject, "error", err)
				ctx := permissions.WithBanLookupError(r.Context(), err)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			if banned {
				w.WriteHeader(stdhttp.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
