package repositories_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestUserRepository_StoreUser(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewUserRepository(sqlHandler)
	user := &domain.User{
		Email:       "foo@example.com",
		DisplayName: "John Doe",
		Password:    "foobar",
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

	{
		dbUser, err := repo.FindByID(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, dbUser, domain.User{
			ID:          user.ID,
			Email:       "foo@example.com",
			DisplayName: "John Smith",
			Password:    "",
			Role:        user.Role,
			Preferences: &domain.Preferences{},
		})
	}

	{
		dbUser, err := repo.FindByEmail("foo@example.com")
		assert.NoError(t, err)
		assert.Equal(t, dbUser, domain.User{
			ID:          user.ID,
			Email:       "foo@example.com",
			DisplayName: "John Smith",
			Password:    "foobar",
			Role:        user.Role,
			Preferences: &domain.Preferences{},
		})
	}
}
