package repositories_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestUserRepository_StoreContest(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestRepository(sqlHandler)
	contest := &domain.Contest{
		Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
		Open:  false,
	}

	{
		err := repo.Store(*contest)
		assert.Nil(t, err)
	}
}

func TestUserRepository_HasOpenContests(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestRepository(sqlHandler)

	{
		hasOpen, err := repo.HasOpenContests()
		assert.Nil(t, err)
		assert.False(t, hasOpen, "no open contests should exist")

		err = repo.Store(domain.Contest{Start: time.Now(), End: time.Now(), Open: true})
		assert.Nil(t, err)

		hasOpen, err = repo.HasOpenContests()
		assert.True(t, hasOpen, "an open contest should exist")
		assert.Nil(t, err)
	}
}
