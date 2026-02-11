package domain

import (
	"context"
	"fmt"

	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PermissionCheckRequest struct {
	Namespace string
	Object    string
	Relation  string
}

type PublicPermissionCheck struct {
	keto      ketoclient.AuthorizationReader
	allowlist PermissionAllowlist
}

func NewPublicPermissionCheck(keto ketoclient.AuthorizationReader, allowlist PermissionAllowlist) *PublicPermissionCheck {
	return &PublicPermissionCheck{keto: keto, allowlist: allowlist}
}

func (s *PublicPermissionCheck) Execute(ctx context.Context, req PermissionCheckRequest) (bool, error) {
	if err := commonroles.RequireAuthenticated(ctx); err != nil {
		return false, err
	}
	subjectID := commonroles.FromContext(ctx).Subject
	if subjectID == "" {
		return false, commondomain.ErrUnauthorized
	}
	if req.Namespace == "" || req.Object == "" || req.Relation == "" {
		return false, fmt.Errorf("%w: namespace, object, and relation are required", commondomain.ErrRequestInvalid)
	}
	if !s.allowlist.Allows(req.Namespace, req.Relation) {
		return false, commondomain.ErrForbidden
	}

	allowed, err := s.keto.CheckPermission(ctx, req.Namespace, req.Object, req.Relation, ketoclient.Subject{ID: subjectID})
	if err != nil {
		return false, fmt.Errorf("%w: check permission failed: %v", commondomain.ErrAuthzUnavailable, err)
	}
	return allowed, nil
}
