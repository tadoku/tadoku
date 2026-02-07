package roles

import (
	"context"
	"fmt"

	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
)

// Service evaluates roles for a given subject ID.
// Subject IDs are expected to be Kratos identity IDs (token "sub").
type Service interface {
	ClaimsForSubject(ctx context.Context, subjectID string) (Claims, error)
}

type KetoService struct {
	keto      ketoclient.PermissionChecker
	namespace string
	object    string
}

func NewKetoService(keto ketoclient.PermissionChecker, namespace, object string) *KetoService {
	return &KetoService{
		keto:      keto,
		namespace: namespace,
		object:    object,
	}
}

func (s *KetoService) ClaimsForSubject(ctx context.Context, subjectID string) (Claims, error) {
	if subjectID == "" || subjectID == "guest" {
		return Claims{Subject: subjectID, Authenticated: false}, nil
	}

	checks := []ketoclient.PermissionCheck{
		{
			Namespace: s.namespace,
			Object:    s.object,
			Relation:  "admins",
			Subject:   ketoclient.Subject{ID: subjectID},
		},
		{
			Namespace: s.namespace,
			Object:    s.object,
			Relation:  "banned",
			Subject:   ketoclient.Subject{ID: subjectID},
		},
	}

	results := s.keto.CheckPermissions(ctx, checks)
	if len(results) != len(checks) {
		err := fmt.Errorf("unexpected keto result count: got=%d want=%d", len(results), len(checks))
		return Claims{Subject: subjectID, Authenticated: true, Err: err}, err
	}

	var (
		adminAllowed  bool
		bannedAllowed bool
	)
	for _, r := range results {
		if r.Err != nil {
			err := fmt.Errorf("keto check %s failed: %w", r.Check.Relation, r.Err)
			return Claims{Subject: subjectID, Authenticated: true, Err: err}, err
		}
		switch r.Check.Relation {
		case "admins":
			adminAllowed = r.Allowed
		case "banned":
			bannedAllowed = r.Allowed
		}
	}

	claims := Claims{
		Subject:       subjectID,
		Authenticated: true,
		Admin:         adminAllowed,
		Banned:        bannedAllowed,
	}
	return claims, nil
}

