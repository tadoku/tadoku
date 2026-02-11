package domain

import (
	"context"
	"fmt"

	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

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
		return false, fmt.Errorf("%w: check permission failed: %v", commondomain.ErrAuthzUnavailable, err)
	}
	return allowed, nil
}
