package domain

import (
	"context"
	"fmt"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
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

func (s *PublicPermissionCheck) Execute(ctx context.Context, subjectID string, req PermissionCheckRequest) (bool, error) {
	if subjectID == "" || subjectID == "guest" {
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
		return false, commondomain.ErrAuthzUnavailable
	}
	return allowed, nil
}

type InternalPermissionCheckRequest struct {
	Namespace string
	Object    string
	Relation  string
	Subject   ketoclient.Subject
}

type InternalPermissionCheck struct {
	keto ketoclient.AuthorizationReader
}

func NewInternalPermissionCheck(keto ketoclient.AuthorizationReader) *InternalPermissionCheck {
	return &InternalPermissionCheck{keto: keto}
}

func (s *InternalPermissionCheck) Execute(ctx context.Context, req InternalPermissionCheckRequest) (bool, error) {
	if req.Namespace == "" || req.Object == "" || req.Relation == "" {
		return false, fmt.Errorf("%w: namespace, object, and relation are required", commondomain.ErrRequestInvalid)
	}
	allowed, err := s.keto.CheckPermission(ctx, req.Namespace, req.Object, req.Relation, req.Subject)
	if err != nil {
		return false, commondomain.ErrAuthzUnavailable
	}
	return allowed, nil
}

