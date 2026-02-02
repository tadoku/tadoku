package repository

import (
	"context"
	"database/sql"
	"errors"

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
