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
		{name: "invalid input", err: &errx.Error{Kind: errx.InvalidInput}, want: stdhttp.StatusBadRequest},
		{name: "unauthorized", err: &errx.Error{Kind: errx.Unauthorized}, want: stdhttp.StatusUnauthorized},
		{name: "forbidden", err: &errx.Error{Kind: errx.Forbidden}, want: stdhttp.StatusForbidden},
		{name: "not found", err: &errx.Error{Kind: errx.NotFound}, want: stdhttp.StatusNotFound},
		{name: "unavailable", err: &errx.Error{Kind: errx.Unavailable}, want: stdhttp.StatusServiceUnavailable},
		{name: "wrapped metadata", err: fmt.Errorf("operation: %w", &errx.Error{Kind: errx.InvalidInput}), want: stdhttp.StatusBadRequest},
		{name: "outer kind wins", err: &errx.Error{Kind: errx.Unavailable, Cause: &errx.Error{Kind: errx.InvalidInput}}, want: stdhttp.StatusServiceUnavailable},
		{name: "unknown outer kind", err: &errx.Error{Cause: &errx.Error{Kind: errx.InvalidInput}}, want: stdhttp.StatusInternalServerError},
		{name: "unknown kind", err: &errx.Error{Kind: errx.Kind(255)}, want: stdhttp.StatusInternalServerError},
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
