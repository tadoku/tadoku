package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// UserExists checks if a user exists in the database.
func (r *Repository) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	// We can use a simple query to check if the user exists
	// Since we don't have a direct UserExists query, we can use FindUserByID or similar
	// For now, let's query the users table directly
	var exists bool
	err := r.psql.QueryRowContext(ctx,
		"select exists(select 1 from users where id = $1)",
		userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not check user existence: %w", err)
	}
	return exists, nil
}
