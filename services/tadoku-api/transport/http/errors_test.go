package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestErrorStatus(t *testing.T) {
	t.Parallel()
	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()

	for _, test := range []struct {
		name string
		ctx  context.Context
		err  error
		want int
	}{
		{name: "internal", err: errx.NewInternalError(""), want: stdhttp.StatusInternalServerError},
		{name: "invalid input", err: errx.NewInvalidInputError(""), want: stdhttp.StatusBadRequest},
		{name: "unauthorized", err: errx.NewUnauthorizedError(""), want: stdhttp.StatusUnauthorized},
		{name: "forbidden", err: errx.NewForbiddenError(""), want: stdhttp.StatusForbidden},
		{name: "not found", err: errx.NewNotFoundError(""), want: stdhttp.StatusNotFound},
		{name: "unavailable", err: errx.NewUnavailableError("", nil), want: stdhttp.StatusServiceUnavailable},
		{name: "deadline exceeded", err: context.DeadlineExceeded, want: stdhttp.StatusGatewayTimeout},
		{name: "wrapped deadline exceeded", err: fmt.Errorf("operation: %w", context.DeadlineExceeded), want: stdhttp.StatusGatewayTimeout},
		{name: "canceled request", ctx: canceledContext, err: context.Canceled, want: 499},
		{name: "internal canceled dependency", err: context.Canceled, want: stdhttp.StatusInternalServerError},
		{name: "unavailable canceled dependency", err: errx.NewUnavailableError("dependency", context.Canceled), want: stdhttp.StatusServiceUnavailable},
		{name: "wrapped metadata", err: fmt.Errorf("operation: %w", errx.NewInvalidInputError("")), want: stdhttp.StatusBadRequest},
		{name: "outer kind wins", err: errx.NewUnavailableError("", errx.NewInvalidInputError("")), want: stdhttp.StatusServiceUnavailable},
		{name: "message is not metadata", err: errors.New("invalid input"), want: stdhttp.StatusInternalServerError},
		{name: "nil", want: stdhttp.StatusInternalServerError},
		{name: "typed nil", err: (*errx.Error)(nil), want: stdhttp.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := test.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			if got := errorStatus(ctx, test.err); got != test.want {
				t.Errorf("errorStatus=%d, want %d", got, test.want)
			}
		})
	}
}

func TestLogOperationErrorClassifiesCancellation(t *testing.T) {
	t.Parallel()
	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()

	for _, test := range []struct {
		name        string
		ctx         context.Context
		err         error
		wantLevel   string
		wantMessage string
	}{
		{name: "live request internal cancellation", ctx: context.Background(), err: context.Canceled, wantLevel: "ERROR", wantMessage: "operation failed"},
		{name: "live request unavailable cancellation", ctx: context.Background(), err: errx.NewUnavailableError("dependency", context.Canceled), wantLevel: "ERROR", wantMessage: "operation failed"},
		{name: "canceled request", ctx: canceledContext, err: context.Canceled, wantLevel: "DEBUG", wantMessage: "operation rejected"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var logs strings.Builder
			server := server{logger: slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))}

			server.logOperationError(test.ctx, "operation", test.err)

			for _, want := range []string{"level=" + test.wantLevel, "msg=\"" + test.wantMessage + "\""} {
				if !strings.Contains(logs.String(), want) {
					t.Errorf("log %q does not contain %q", logs.String(), want)
				}
			}
		})
	}
}
