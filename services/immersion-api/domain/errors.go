package domain

import (
	"errors"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// Common errors
var (
	ErrRequestInvalid = errors.New("request is invalid")
	ErrNotFound       = errors.New("not found")
	ErrForbidden      = commondomain.ErrForbidden
	ErrAuthzUnavailable = commondomain.ErrAuthzUnavailable
	ErrUnauthorized   = errors.New("unauthorized")
	ErrConflict       = errors.New("conflict")
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
