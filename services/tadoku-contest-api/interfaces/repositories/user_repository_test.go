package repositories_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/tadoku-contest-api/domain"
	"github.com/tadoku/tadoku/services/tadoku-contest-api/interfaces/repositories"
)

func TestUserRepository_StoreUser(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewUserRepository(sqlHandler)
	user := &domain.User{
		Email:       "foo@example.com",
		DisplayName: "John Doe",
		Role:        domain.RoleUser,
		Preferences: &domain.Preferences{},
	}

	{
		err := repo.Store(user)
		assert.NoError(t, err)
	}

	{
		user.DisplayName = "John Smith"
		err := repo.Store(user)
		assert.NoError(t, err)
	}
}
