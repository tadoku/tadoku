package domain

import "errors"

// Common errors
var (
	ErrRequestInvalid = errors.New("request is invalid")
	ErrNotFound       = errors.New("not found")
	ErrForbidden      = errors.New("not allowed")
	ErrUnauthorized   = errors.New("unauthorized")
)

// Log errors
var (
	ErrInvalidLog = errors.New("unable to validate log")
)

// Contest errors
var (
	ErrInvalidContest             = errors.New("unable to validate contest")
	ErrInvalidContestRegistration = errors.New("language selection is not valid for contest")
)
