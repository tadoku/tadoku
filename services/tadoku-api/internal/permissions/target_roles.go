package permissions

import (
	"context"
	"fmt"

	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
)

type KetoService struct {
	keto ketoclient.AuthorizationReader
}

func NewKetoService(keto ketoclient.AuthorizationReader) *KetoService {
	return &KetoService{keto: keto}
}

func (s *KetoService) RolesForSubject(ctx context.Context, subjectID string) (TargetRoles, error) {
	object, err := rolesObject(ctx)
	if err != nil {
		return TargetRoles{}, err
	}

	if subjectID == "" || subjectID == "guest" {
		return TargetRoles{}, nil
	}

	checks := []ketoclient.PermissionCheck{
		{
			Namespace: "app",
			Object:    object,
			Relation:  "admin",
			Subject:   ketoclient.Subject{ID: subjectID},
		},
		{
			Namespace: "app",
			Object:    object,
			Relation:  "is_banned",
			Subject:   ketoclient.Subject{ID: subjectID},
		},
	}

	results := s.keto.CheckPermissions(ctx, checks)
	if len(results) != len(checks) {
		err := fmt.Errorf("unexpected keto result count: got=%d want=%d", len(results), len(checks))
		return TargetRoles{}, err
	}

	var (
		adminAllowed  bool
		bannedAllowed bool
	)
	for _, r := range results {
		if r.Err != nil {
			err := fmt.Errorf("keto check %s failed: %w", r.Check.Relation, r.Err)
			return TargetRoles{}, err
		}
		switch r.Check.Relation {
		case "admin":
			adminAllowed = r.Allowed
		case "is_banned":
			bannedAllowed = r.Allowed
		}
	}

	return TargetRoles{Admin: adminAllowed, Banned: bannedAllowed}, nil
}

func (s *KetoService) RolesForSubjects(ctx context.Context, subjectIDs []string) (map[string]TargetRoles, error) {
	object, err := rolesObject(ctx)
	if err != nil {
		return nil, err
	}

	out := make(map[string]TargetRoles, len(subjectIDs))
	unique := make(map[string]struct{}, len(subjectIDs))

	for _, subjectID := range subjectIDs {
		if subjectID == "" || subjectID == "guest" {
			out[subjectID] = TargetRoles{}
			continue
		}
		unique[subjectID] = struct{}{}
	}
	if len(unique) == 0 {
		return out, nil
	}

	adminIDs, err := s.keto.ListSubjectIDsForRelation(ctx, "app", object, "admins")
	if err != nil {
		return nil, fmt.Errorf("keto list admins failed: %w", err)
	}
	bannedIDs, err := s.keto.ListSubjectIDsForRelation(ctx, "app", object, "banned")
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
		out[subjectID] = TargetRoles{Admin: admin, Banned: banned}
	}

	return out, nil
}
