package http

import (
	"context"
	stdhttp "net/http"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var errInvalidUUID = errx.NewInvalidInputError("invalid UUID")

func (s *server) logOperationError(ctx context.Context, operation string, err error) {
	if errorStatus(err) >= stdhttp.StatusInternalServerError {
		s.logger.ErrorContext(ctx, operation+" failed", "error", err)
		return
	}
	s.logger.DebugContext(ctx, operation+" rejected", "error", err)
}

func errorStatus(err error) int {
	switch errx.KindOf(err) {
	case errx.InvalidInput:
		return stdhttp.StatusBadRequest
	case errx.Unauthorized:
		return stdhttp.StatusUnauthorized
	case errx.Forbidden:
		return stdhttp.StatusForbidden
	case errx.NotFound:
		return stdhttp.StatusNotFound
	case errx.Unavailable:
		return stdhttp.StatusServiceUnavailable
	default:
		return stdhttp.StatusInternalServerError
	}
}
