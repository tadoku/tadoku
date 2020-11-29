package repositories_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestContestLogRepository_StoreUpdateDeleteLog(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestLogRepository(sqlHandler)
	log := &domain.ContestLog{
		ContestID:   1,
		UserID:      1,
		Language:    domain.Japanese,
		Amount:      10,
		MediumID:    1,
		Description: "foobar",
	}

	{
		err := repo.Store(log)
		assert.NoError(t, err)
	}

	{
		updatedLog := &domain.ContestLog{
			ID:          log.ID,
			ContestID:   1,
			UserID:      1,
			Language:    domain.Korean,
			Amount:      20,
			MediumID:    2,
			Description: "foobar 2",
		}
		assert.NotEqual(t, 0, updatedLog.ID)
		err := repo.Store(updatedLog)
		assert.NoError(t, err)
	}

	{
		err := repo.Delete(log.ID)
		assert.NoError(t, err)

		_, err = repo.FindByID(log.ID)
		assert.EqualError(t, err, domain.ErrNotFound.Error())
	}
}

func TestContestLogRepository_FindAllByContestAndUser(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestLogRepository(sqlHandler)

	contestID := uint64(1)
	userID := uint64(1)

	expected := []struct {
		language    domain.LanguageCode
		medium      domain.MediumID
		amount      float32
		description string
	}{
		{domain.Japanese, domain.MediumBook, 10, "foobar"},
		{domain.Korean, domain.MediumComic, 20, "foobar 2"},
		{domain.Global, domain.MediumNet, 30, "foobar 3"},
	}

	// Correct logs
	{
		for _, data := range expected {
			log := &domain.ContestLog{
				ContestID:   contestID,
				UserID:      userID,
				Language:    data.language,
				MediumID:    data.medium,
				Amount:      data.amount,
				Description: data.description,
			}

			err := repo.Store(log)
			assert.NoError(t, err)
		}
	}

	// Create unrelated rankings to check if it is really working
	{
		for _, language := range []domain.LanguageCode{domain.Korean, domain.Global} {
			log := &domain.ContestLog{
				ContestID:   contestID + 1,
				UserID:      userID,
				Language:    language,
				MediumID:    domain.MediumBook,
				Amount:      0,
				Description: "barbar",
			}

			err := repo.Store(log)
			assert.NoError(t, err)
		}
	}

	logs, err := repo.FindAll(contestID, userID)
	assert.NoError(t, err)

	for _, expected := range expected {
		var log domain.ContestLog
		for _, l := range logs {
			if l.Language == expected.language {
				log = l
			}
		}

		assert.Equal(t, expected.amount, log.Amount)
		assert.Equal(t, expected.medium, log.MediumID)
		assert.Equal(t, expected.description, log.Description)
		assert.Equal(t, contestID, log.ContestID)
		assert.Equal(t, userID, log.UserID)
	}
}

func TestContestLogRepository_FindRecent(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewContestLogRepository(sqlHandler)
	userRepo := repositories.NewUserRepository(sqlHandler)

	contestID := uint64(1)
	user := &domain.User{
		Email:       "foo@example.com",
		DisplayName: "John Smith",
		Password:    "foobar",
		Role:        domain.RoleUser,
		Preferences: &domain.Preferences{},
	}

	expected := []struct {
		language    domain.LanguageCode
		medium      domain.MediumID
		amount      float32
		description string
	}{
		{domain.Japanese, domain.MediumBook, 10, "foobar"},
		{domain.Korean, domain.MediumComic, 20, "foobar 2"},
		{domain.Croatian, domain.MediumGame, 30, "foobar 3"},
		{domain.Dutch, domain.MediumSentences, 40, "foobar 4"},
	}

	// Create user
	{
		err := userRepo.Store(user)
		assert.NoError(t, err)
	}

	// Correct logs
	{
		for _, data := range expected {
			log := &domain.ContestLog{
				ContestID:   contestID,
				UserID:      user.ID,
				Language:    data.language,
				MediumID:    data.medium,
				Amount:      data.amount,
				Description: data.description,
			}

			err := repo.Store(log)
			assert.NoError(t, err)
		}
	}

	// Create unrelated logs to check if it is really working
	{
		for _, language := range []domain.LanguageCode{domain.Korean, domain.Japanese} {
			log := &domain.ContestLog{
				ContestID:   contestID + 1,
				UserID:      user.ID,
				Language:    language,
				MediumID:    domain.MediumBook,
				Amount:      0,
				Description: "barbar",
			}

			err := repo.Store(log)
			assert.NoError(t, err)
		}
	}

	// Test when there are too few logs
	{
		logs, err := repo.FindRecent(contestID, 25)
		assert.NoError(t, err)

		count := len(expected)
		for i, expected := range expected {
			log := logs[count-i-1]

			assert.Equal(t, expected.amount, log.Amount)
			assert.Equal(t, expected.medium, log.MediumID)
			assert.Equal(t, expected.description, log.Description)
			assert.Equal(t, contestID, log.ContestID)
			assert.Equal(t, user.DisplayName, log.UserDisplayName)
		}
	}

	// Test when there are too many logs
	{
		logs, err := repo.FindRecent(contestID, 2)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(logs))
		assert.Equal(t, expected[len(expected)-1].description, logs[0].Description)
	}
}
