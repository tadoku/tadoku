package domain

import "errors"

var (
	ErrForbidden        = errors.New("not allowed")
	ErrAuthzUnavailable = errors.New("authorization unavailable")

	ErrRequestInvalid = errors.New("request is invalid")
	ErrNotFound       = errors.New("not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrConflict       = errors.New("conflict")
)
