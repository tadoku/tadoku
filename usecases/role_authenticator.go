package usecases

import (
	"github.com/srvc/fail"

	"github.com/tadoku/api/domain"
)

var (
	// ErrRoleAuthenticatorMissingUser there is no user to check against
	ErrRoleAuthenticatorMissingUser = fail.Errorf("user is missing, nobody to check against")
	// ErrInsufficientAccess is triggered when a user has a lower role than required
	ErrInsufficientAccess = fail.Errorf("user has insufficient privileges")
)

// RoleAuthenticator contains all business logic for accessing resources
type RoleAuthenticator interface {
	IsAllowed(user *domain.User, minRole domain.Role) error
}

type roleAuthenticator struct{}

// NewRoleAuthenticator instantiates a new role authenticator
func NewRoleAuthenticator() RoleAuthenticator {
	return &roleAuthenticator{}
}

// IsAllowed checks if a user has sufficient privileges to access a resource
func (ra *roleAuthenticator) IsAllowed(user *domain.User, minRole domain.Role) error {
	if minRole <= domain.RoleGuest {
		return nil
	}

	if user == nil {
		return ErrRoleAuthenticatorMissingUser
	}

	if user.Role < minRole {
		return ErrInsufficientAccess
	}

	return nil
}
