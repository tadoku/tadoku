package http

import (
	"context"
	"log/slog"
	stdhttp "net/http"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

const authzRoleGetPattern = "GET /authz/current-user/role"

func RejectBannedUsers(
	admit func(context.Context) permissions.Admission,
	logger *slog.Logger,
) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			switch admission := admit(r.Context()).(type) {
			case permissions.Admitted:
			case permissions.AdmittedBanUnknown:
				user := identity.FromContext(r.Context())
				logger.ErrorContext(r.Context(), "banned-user check unavailable; allowing request",
					"subject", user.Subject, "error", admission.Err,
				)
				ctx := permissions.WithBanState(r.Context(), permissions.BanUnknown(admission.Err))
				r = r.WithContext(ctx)
			case permissions.Banned:
				if r.Pattern != authzRoleGetPattern {
					w.WriteHeader(stdhttp.StatusForbidden)
					return
				}
				ctx := permissions.WithBanState(r.Context(), permissions.ConfirmedBan())
				r = r.WithContext(ctx)
			case permissions.NoAccess:
				w.WriteHeader(stdhttp.StatusForbidden)
				return
			case permissions.Unavailable:
				logger.ErrorContext(r.Context(), "tenant admission unavailable", "error", admission.Err)
				w.WriteHeader(stdhttp.StatusServiceUnavailable)
				return
			default:
				w.WriteHeader(stdhttp.StatusServiceUnavailable)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
