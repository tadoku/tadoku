package repositories_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestContestRepository_StoreContest(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestRepository(sqlHandler)
	contest := &domain.Contest{
		Description: "Round 2019-05",
		Start:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		End:         time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
		Open:        false,
	}

	{
		err := repo.Store(contest)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, contest.ID)
	}

	{
		updatedContest := &domain.Contest{
			ID:          contest.ID,
			Description: "Round 2019-01",
			Start:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			End:         time.Date(2019, 1, 30, 0, 0, 0, 0, time.UTC),
			Open:        false,
		}
		err := repo.Store(updatedContest)
		assert.NoError(t, err)
	}
}

func TestContestRepository_GetOpenContests(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestRepository(sqlHandler)

	{
		ids, err := repo.GetOpenContests()
		assert.NoError(t, err)
		assert.Empty(t, ids, "no open contests should exist")

		err = repo.Store(&domain.Contest{Start: time.Now(), End: time.Now(), Open: true})
		assert.NoError(t, err)

		ids, err = repo.GetOpenContests()
		assert.Equal(t, 1, len(ids), "an open contest should exist")
		assert.NoError(t, err)
	}
}

func TestContestRepository_GetRunningContests(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestRepository(sqlHandler)

	{
		ids, err := repo.GetRunningContests()
		assert.NoError(t, err)
		assert.Empty(t, ids, "no running contests should exist")

		for _, contest := range []*domain.Contest{
			{Start: time.Now().Add(-1 * time.Hour), End: time.Now().Add(1 * time.Hour), Open: true},
			{Start: time.Now().Add(-1 * time.Hour), End: time.Now().Add(1 * time.Hour), Open: false},
			{Start: time.Now().Add(-5 * time.Hour), End: time.Now().Add(-1 * time.Hour), Open: false},
		} {
			err = repo.Store(contest)
			assert.NoError(t, err, "saving seed contest should return no error")
		}

		ids, err = repo.GetOpenContests()
		assert.Equal(t, 1, len(ids), "only one running contest should exist")
		assert.NoError(t, err)
	}
}

func TestContestRepository_FindLatest(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestRepository(sqlHandler)

	{
		contest, err := repo.FindLatest()
		assert.EqualError(t, err, domain.ErrNotFound.Error())
		assert.Empty(t, contest, "no contests should be found")

		expected := domain.Contest{Description: "Foo 2019", Start: time.Now(), End: time.Now(), Open: true}
		err = repo.Store(&expected)
		assert.NoError(t, err)

		contest, err = repo.FindLatest()
		assert.Equal(t, expected.Description, contest.Description, "contest should have the same description")
		assert.Equal(t, expected.Open, contest.Open, "contest should both be open")
		assert.NoError(t, err)
	}
}

func TestContestRepository_FindByID(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestRepository(sqlHandler)

	{
		contest, err := repo.FindByID(0)
		assert.EqualError(t, err, domain.ErrNotFound.Error())
		assert.Empty(t, contest, "no contests should be found")

		expected := domain.Contest{Description: "Foo 2019", Start: time.Now(), End: time.Now(), Open: true}
		err = repo.Store(&expected)
		assert.NoError(t, err)

		contest, err = repo.FindByID(expected.ID)
		assert.Equal(t, expected.Description, contest.Description, "contest should have the same description")
		assert.Equal(t, expected.Open, contest.Open, "contest should both be open")
		assert.NoError(t, err)
	}
}
