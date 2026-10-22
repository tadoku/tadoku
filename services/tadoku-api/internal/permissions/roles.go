package permissions

import (
	"context"
	"maps"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Roles reads the stored admins and banned relations on the context tenant's app:root, without inheritance.
// A subject that is not a UUID, such as "" or "guest", has no roles.
type Roles struct {
	store store
}

func NewRoles(db *pgxpool.Pool) *Roles {
	return &Roles{store: store{db: db}}
}

func (r *Roles) RolesForSubject(ctx context.Context, subjectID string) (TargetRoles, error) {
	roles, err := r.RolesForSubjects(ctx, []string{subjectID})
	if err != nil {
		return TargetRoles{}, err
	}
	return roles[subjectID], nil
}

func (r *Roles) RolesForSubjects(ctx context.Context, subjectIDs []string) (map[string]TargetRoles, error) {
	users := make(map[string]uuid.UUID, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		if id, err := uuid.Parse(subjectID); err == nil {
			users[subjectID] = id
		}
	}

	stored, err := r.store.appRelationships(ctx, slices.Collect(maps.Values(users)))
	if err != nil {
		return nil, err
	}

	roles := make(map[string]TargetRoles, len(subjectIDs))
	for _, subjectID := range subjectIDs {
		roles[subjectID] = TargetRoles{}
	}
	for subjectID, id := range users {
		roles[subjectID] = TargetRoles{Admin: stored[Admin(id)], Banned: stored[Ban(id)]}
	}
	return roles, nil
}
