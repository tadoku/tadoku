package domain

import "errors"

var (
	// ErrForbidden is returned when the caller is not allowed to perform an action.
	ErrForbidden = errors.New("not allowed")
	// ErrAuthzUnavailable is returned when we cannot evaluate authorization (e.g. Keto unavailable).
	ErrAuthzUnavailable = errors.New("authorization unavailable")

	// ErrRequestInvalid is returned when the caller's input is invalid.
	ErrRequestInvalid = errors.New("request is invalid")
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")
	// ErrUnauthorized is returned when the caller is not authenticated.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrConflict is returned when the request conflicts with an existing resource.
	ErrConflict = errors.New("conflict")
)
