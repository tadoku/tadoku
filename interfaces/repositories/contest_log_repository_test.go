package repositories_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestContestLogRepository_StoreContest(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestLogRepository(sqlHandler)
	log := &domain.ContestLog{
		ContestID: 1,
		UserID:    1,
		Language:  domain.Japanese,
		Amount:    10,
		MediumID:  1,
	}

	{
		err := repo.Store(*log)
		assert.NoError(t, err)
	}

	{
		updatedLog := &domain.ContestLog{
			ID:        1,
			ContestID: 1,
			UserID:    1,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  1,
		}
		err := repo.Store(*updatedLog)
		assert.EqualError(t, err, "not yet implemented")
	}
}
