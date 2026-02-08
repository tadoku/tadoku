package domain

import (
	"context"

	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type RoleGet struct {
	roles commonroles.Service
}

func NewRoleGet(roles commonroles.Service) *RoleGet {
	return &RoleGet{roles: roles}
}

// Execute returns one of: guest, user, admin, banned.
func (s *RoleGet) Execute(ctx context.Context, subjectID string) (string, error) {
	if subjectID == "" || subjectID == "guest" {
		return "guest", nil
	}

	claims, err := s.roles.ClaimsForSubject(ctx, subjectID)
	if err != nil {
		return "", commondomain.ErrAuthzUnavailable
	}
	if claims.Banned {
		return "banned", nil
	}
	if claims.Admin {
		return "admin", nil
	}
	return "user", nil
}

