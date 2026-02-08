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
	ClaimsForSubjects(ctx context.Context, subjectIDs []string) (map[string]Claims, error)
}

type KetoService struct {
	keto      ketoclient.AuthorizationReader
	namespace string
	object    string
}

func NewKetoService(keto ketoclient.AuthorizationReader, namespace, object string) *KetoService {
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

func (s *KetoService) ClaimsForSubjects(ctx context.Context, subjectIDs []string) (map[string]Claims, error) {
	out := make(map[string]Claims, len(subjectIDs))
	unique := make(map[string]struct{}, len(subjectIDs))

	for _, subjectID := range subjectIDs {
		if subjectID == "" || subjectID == "guest" {
			out[subjectID] = Claims{Subject: subjectID, Authenticated: false}
			continue
		}
		unique[subjectID] = struct{}{}
	}
	if len(unique) == 0 {
		return out, nil
	}

	adminIDs, err := s.keto.ListSubjectIDsForRelation(ctx, s.namespace, s.object, "admins")
	if err != nil {
		return nil, fmt.Errorf("keto list admins failed: %w", err)
	}
	bannedIDs, err := s.keto.ListSubjectIDsForRelation(ctx, s.namespace, s.object, "banned")
	if err != nil {
		return nil, fmt.Errorf("keto list banned failed: %w", err)
	}

	adminSet := make(map[string]struct{}, len(adminIDs))
	for _, id := range adminIDs {
		adminSet[id] = struct{}{}
	}

	bannedSet := make(map[string]struct{}, len(bannedIDs))
	for _, id := range bannedIDs {
		bannedSet[id] = struct{}{}
	}

	for subjectID := range unique {
		_, admin := adminSet[subjectID]
		_, banned := bannedSet[subjectID]
		out[subjectID] = Claims{
			Subject:       subjectID,
			Authenticated: true,
			Admin:         admin,
			Banned:        banned,
		}
	}

	return out, nil
}
