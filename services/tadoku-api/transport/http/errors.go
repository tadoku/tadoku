package http

import (
	"context"
	"errors"
	stdhttp "net/http"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/app/api"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var errInvalidUUID = errx.NewInvalidInputError("invalid UUID")

func (s *server) logOperationError(ctx context.Context, operation string, err error) {
	if errorStatus(ctx, err) >= stdhttp.StatusInternalServerError {
		s.logger.ErrorContext(ctx, operation+" failed", "error", err)
		return
	}
	s.logger.DebugContext(ctx, operation+" rejected", "error", err)
}

func writeResponseError(w stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
	var identityError *api.IdentityError
	switch {
	case errors.As(err, &identityError):
		writeJSON(w, errorStatus(request.Context(), err), map[string]string{
			"code":    identityError.Code,
			"message": identityError.Error(),
		})
	case strings.HasPrefix(request.URL.Path, "/identity/") && errx.KindOf(err) == errx.Unavailable:
		writeJSON(w, errorStatus(request.Context(), err), map[string]string{"message": "identity service unavailable"})
	case strings.HasPrefix(request.URL.Path, "/identity/") && errx.KindOf(err) == errx.InvalidInput:
		writeJSON(w, errorStatus(request.Context(), err), map[string]string{"message": err.Error()})
	case errx.KindOf(err) == errx.NotImplemented:
		writeJSON(w, stdhttp.StatusNotImplemented, map[string]string{"message": "not implemented"})
	case errors.Is(err, api.ErrAccountDeletionInProgress):
		writeJSON(w, stdhttp.StatusConflict, openapi.ImmersionAccountDeletionInProgress{
			Error: openapi.AccountDeletionInProgress,
		})
	case errx.KindOf(err) == errx.Internal:
		writeJSON(w, stdhttp.StatusInternalServerError, map[string]string{"message": "Internal Server Error"})
	default:
		w.WriteHeader(errorStatus(request.Context(), err))
	}
}

func writeRequestError(w stdhttp.ResponseWriter, err error) {
	var tooLarge *stdhttp.MaxBytesError
	if errors.As(err, &tooLarge) {
		w.WriteHeader(stdhttp.StatusRequestEntityTooLarge)
		return
	}

	w.WriteHeader(stdhttp.StatusBadRequest)
}

func errorStatus(ctx context.Context, err error) int {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return stdhttp.StatusGatewayTimeout
	case errors.Is(err, context.Canceled) && errors.Is(ctx.Err(), context.Canceled):
		return 499
	}

	switch errx.KindOf(err) {
	case errx.Internal:
		return stdhttp.StatusInternalServerError
	case errx.NotImplemented:
		return stdhttp.StatusNotImplemented
	case errx.Throttled:
		return stdhttp.StatusTooManyRequests
	case errx.InvalidInput:
		return stdhttp.StatusBadRequest
	case errx.Unauthorized:
		return stdhttp.StatusUnauthorized
	case errx.Forbidden:
		return stdhttp.StatusForbidden
	case errx.NotFound:
		return stdhttp.StatusNotFound
	case errx.Conflict:
		return stdhttp.StatusConflict
	case errx.Unavailable:
		return stdhttp.StatusServiceUnavailable
	default:
		return stdhttp.StatusInternalServerError
	}
}
