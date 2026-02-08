package domain

import (
	"errors"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// Common errors
var (
	ErrRequestInvalid   = commondomain.ErrRequestInvalid
	ErrNotFound         = commondomain.ErrNotFound
	ErrForbidden        = commondomain.ErrForbidden
	ErrAuthzUnavailable = commondomain.ErrAuthzUnavailable
	ErrUnauthorized     = commondomain.ErrUnauthorized
	ErrConflict         = commondomain.ErrConflict
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
