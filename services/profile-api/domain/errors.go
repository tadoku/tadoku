package domain

import "errors"

// Profile errors
var (
	ErrProfileAlreadyExists = errors.New("profile for user already exists")
	ErrProfileNotFound      = errors.New("profile not found")
	ErrInvalidProfile       = errors.New("unable to validate profile")
)

// Common errors
var (
	ErrForbidden      = errors.New("not allowed")
	ErrRequestInvalid = errors.New("request is invalid")
)
