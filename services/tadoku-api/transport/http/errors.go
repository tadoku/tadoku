package http

import (
	"context"
	"errors"
	stdhttp "net/http"

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
