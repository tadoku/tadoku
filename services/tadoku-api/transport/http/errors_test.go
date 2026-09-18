package http

import (
	"errors"
	"fmt"
	stdhttp "net/http"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestErrorStatus(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid input", err: errx.NewInvalidInputError(""), want: stdhttp.StatusBadRequest},
		{name: "unauthorized", err: errx.NewUnauthorizedError(""), want: stdhttp.StatusUnauthorized},
		{name: "forbidden", err: errx.NewForbiddenError(""), want: stdhttp.StatusForbidden},
		{name: "not found", err: errx.NewNotFoundError(""), want: stdhttp.StatusNotFound},
		{name: "unavailable", err: errx.NewUnavailableError("", nil), want: stdhttp.StatusServiceUnavailable},
		{name: "wrapped metadata", err: fmt.Errorf("operation: %w", errx.NewInvalidInputError("")), want: stdhttp.StatusBadRequest},
		{name: "outer kind wins", err: errx.NewUnavailableError("", errx.NewInvalidInputError("")), want: stdhttp.StatusServiceUnavailable},
		{name: "message is not metadata", err: errors.New("invalid input"), want: stdhttp.StatusInternalServerError},
		{name: "nil", want: stdhttp.StatusInternalServerError},
		{name: "typed nil", err: (*errx.Error)(nil), want: stdhttp.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := errorStatus(test.err); got != test.want {
				t.Errorf("errorStatus=%d, want %d", got, test.want)
			}
		})
	}
}
