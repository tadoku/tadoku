package domain

import (
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// Common errors
var (
	ErrForbidden        = commondomain.ErrForbidden
	ErrAuthzUnavailable = commondomain.ErrAuthzUnavailable
	ErrRequestInvalid   = commondomain.ErrRequestInvalid
	ErrUnauthorized     = commondomain.ErrUnauthorized
)
