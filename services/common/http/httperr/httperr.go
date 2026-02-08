package httperr

import (
	"errors"
	"net/http"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// StatusCode maps common domain sentinel errors to HTTP status codes.
// It returns (code, true) when a mapping exists, otherwise (0, false).
func StatusCode(err error) (int, bool) {
	switch {
	case errors.Is(err, commondomain.ErrRequestInvalid):
		return http.StatusBadRequest, true
	case errors.Is(err, commondomain.ErrUnauthorized):
		return http.StatusUnauthorized, true
	case errors.Is(err, commondomain.ErrForbidden):
		return http.StatusForbidden, true
	case errors.Is(err, commondomain.ErrNotFound):
		return http.StatusNotFound, true
	case errors.Is(err, commondomain.ErrConflict):
		return http.StatusConflict, true
	case errors.Is(err, commondomain.ErrAuthzUnavailable):
		return http.StatusServiceUnavailable, true
	default:
		return 0, false
	}
}
