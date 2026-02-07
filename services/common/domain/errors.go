package domain

import "errors"

var (
	// ErrForbidden is returned when the caller is not allowed to perform an action.
	ErrForbidden = errors.New("not allowed")
	// ErrAuthzUnavailable is returned when we cannot evaluate authorization (e.g. Keto unavailable).
	ErrAuthzUnavailable = errors.New("authorization unavailable")
)

