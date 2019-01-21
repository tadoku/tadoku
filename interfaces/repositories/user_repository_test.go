package repositories_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestUserRepository_StoreUser(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	passwordHasher := interfaces.NewMockHasher(ctrl)
	passwordHasher.EXPECT().Hash("foobar").Return("ABCDEFG", nil)

	repo := repositories.NewUserRepository(sqlHandler, passwordHasher)
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
