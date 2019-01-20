package repositories_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestUserRepository_StoreNewUser(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewUserRepository(sqlHandler)
	user := domain.User{
		Email:       "foo@example.com",
		DisplayName: "John Doe",
		Password:    "foobar",
		Role:        domain.RoleUser,
		Preferences: &domain.Preferences{},
	}

	err := repo.Store(user)
	assert.Nil(t, err)
}
