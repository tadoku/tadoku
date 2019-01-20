package repositories_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestUserRepository_StoreUser(t *testing.T) {
	t.Parallel()
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
		err := repo.Store(*user)
		assert.Nil(t, err)
	}

	{
		user.ID = 1
		user.DisplayName = "John Smith"
		err := repo.Store(*user)
		assert.Nil(t, err)
	}

	{
		dbUser, err := repo.FindByID(1)
		assert.Nil(t, err)
		assert.Equal(t, dbUser, domain.User{
			ID:          1,
			Email:       "foo@example.com",
			DisplayName: "John Smith",
			Password:    "",
			Role:        user.Role,
			Preferences: &domain.Preferences{},
		})
	}
}
