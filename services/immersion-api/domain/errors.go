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
	ErrAccountDeletionNotLocked  = fmt.Errorf("account deletion is not locked: %w", commondomain.ErrConflict)
	ErrRunningContestOwned       = fmt.Errorf("account owns a running contest: %w", commondomain.ErrConflict)
	ErrFeatureAccessUnavailable  = errors.New("feature access unavailable")
)

// Log errors
var (
	ErrInvalidLog             = errors.New("unable to validate log")
	ErrInvalidScoringRuleSet  = errors.New("invalid scoring rule set")
	ErrScoringRuleSetNotFound = errors.New("scoring rule set not found")
	ErrInvalidTags            = errors.New("invalid tags")
	ErrLogFrozen              = fmt.Errorf("log is frozen: %w", commondomain.ErrConflict)
)

// Contest errors
var (
	ErrInvalidContest             = errors.New("unable to validate contest")
	ErrInvalidContestRegistration = errors.New("language selection is not valid for contest")
)
