package domain

import "errors"

// Common errors
var (
	ErrForbidden      = errors.New("not allowed")
	ErrRequestInvalid = errors.New("request is invalid")
)
