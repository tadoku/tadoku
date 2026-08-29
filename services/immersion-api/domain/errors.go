package domain

import (
	"errors"
	"fmt"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// Common errors
var (
	ErrRequestInvalid            = commondomain.ErrRequestInvalid
	ErrNotFound                  = commondomain.ErrNotFound
	ErrForbidden                 = commondomain.ErrForbidden
	ErrAuthzUnavailable          = commondomain.ErrAuthzUnavailable
	ErrUnauthorized              = commondomain.ErrUnauthorized
	ErrConflict                  = commondomain.ErrConflict
	ErrAccountDeletionInProgress = fmt.Errorf("account deletion in progress: %w", commondomain.ErrConflict)
)

// Log errors
var (
	ErrInvalidLog             = errors.New("unable to validate log")
	ErrInvalidScoringRuleSet  = errors.New("invalid scoring rule set")
	ErrScoringRuleSetNotFound = errors.New("scoring rule set not found")
	ErrInvalidTags            = errors.New("invalid tags")
)

// Contest errors
var (
	ErrInvalidContest             = errors.New("unable to validate contest")
	ErrInvalidContestRegistration = errors.New("language selection is not valid for contest")
)
