package http

import (
	stdhttp "net/http"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

var errInvalidUUID = &errx.Error{
	Kind:    errx.InvalidInput,
	Message: "invalid UUID",
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
