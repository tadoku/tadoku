package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// GetUserRole returns the role for a user by their ID (as string).
// Returns "user" if no specific role is found.
// This implements the DatabaseRoleRepository interface from the middleware.
func (r *Repository) GetUserRole(ctx context.Context, userID string) (string, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "user", nil // Invalid UUID means no role found
	}

	role, err := r.q.GetUserRoleByUserID(ctx, uid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "user", nil
		}
		return "", err
	}
	return role, nil
}

// GetAllUserRoles returns a map of user_id -> role for all users with special roles.
// Users not in the map have the default "user" role.
func (r *Repository) GetAllUserRoles(ctx context.Context) (map[string]string, error) {
	rows, err := r.q.ListAllUserRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not list user roles: %w", err)
	}

	roleMap := make(map[string]string, len(rows))
	for _, row := range rows {
		roleMap[row.UserID.String()] = row.Role
	}
	return roleMap, nil
}
