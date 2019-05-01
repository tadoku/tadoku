package repositories_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestContestRepository_StoreContest(t *testing.T) {
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
		assert.NoError(t, err)
	}

	{
		updatedContest := &domain.Contest{
			ID:    1,
			Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2019, 1, 30, 0, 0, 0, 0, time.UTC),
			Open:  false,
		}
		err := repo.Store(*updatedContest)
		assert.NoError(t, err)
	}
}

func TestContestRepository_GetOpenContests(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestRepository(sqlHandler)

	{
		ids, err := repo.GetOpenContests()
		assert.NoError(t, err)
		assert.Empty(t, ids, "no open contests should exist")

		err = repo.Store(domain.Contest{Start: time.Now(), End: time.Now(), Open: true})
		assert.NoError(t, err)

		ids, err = repo.GetOpenContests()
		assert.Equal(t, 1, len(ids), "an open contest should exist")
		assert.NoError(t, err)
	}
}
