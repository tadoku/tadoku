package repositories_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestRankingRepository_StoreRanking(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)
	ranking := &domain.Ranking{
		ID:        1,
		ContestID: 1,
		UserID:    1,
		Language:  domain.Japanese,
		Amount:    0,
		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	{
		err := repo.Store(*ranking)
		assert.NoError(t, err)
	}

	{
		updatedRanking := &domain.Ranking{
			ID:     1,
			Amount: 2,
		}
		err := repo.Store(*updatedRanking)
		assert.NoError(t, err)
	}
}

func TestRankingRepository_GetAllLanguagesForContestAndUser(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)
	rankingJapanese := &domain.Ranking{
		ContestID: 1,
		UserID:    1,
		Language:  domain.Japanese,
		Amount:    0,
		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	rankingChinese := &domain.Ranking{
		ContestID: 1,
		UserID:    1,
		Language:  domain.Chinese,
		Amount:    0,
		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	rankingGlobal := &domain.Ranking{
		ContestID: 1,
		UserID:    1,
		Language:  domain.Global,
		Amount:    0,
		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	rankingSingleLanguage := &domain.Ranking{
		ContestID: 1,
		UserID:    2,
		Language:  domain.Chinese,
		Amount:    0,
		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	{
		for _, r := range []*domain.Ranking{rankingJapanese, rankingChinese, rankingGlobal, rankingSingleLanguage} {
			err := repo.Store(*r)
			assert.NoError(t, err)
		}
	}

	{
		languages, err := repo.GetAllLanguagesForContestAndUser(1, 1)
		assert.NoError(t, err)
		assert.Equal(t, len(languages), 2)
		assert.Equal(t, languages[0], domain.Japanese)
		assert.Equal(t, languages[1], domain.Chinese)
	}

	{
		languages, err := repo.GetAllLanguagesForContestAndUser(1, 2)
		assert.NoError(t, err)
		assert.Equal(t, len(languages), 1)
		assert.Equal(t, languages[0], domain.Chinese)
	}
}

func TestRankingRepository_RankingsForContest(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)

	contestID := uint64(1)

	type testCase struct {
		contestID uint64
		userID    uint64
		language  domain.LanguageCode
		amount    float32
	}
	expected := []testCase{
		{contestID, 3, domain.Global, 30},
		{contestID, 2, domain.Global, 20},
		{contestID, 1, domain.Global, 10},
	}

	// Correct rankings
	{
		for _, data := range []testCase{expected[2], expected[1], expected[0]} {
			ranking := &domain.Ranking{
				ContestID: data.contestID,
				UserID:    data.userID,
				Language:  data.language,
				Amount:    data.amount,
			}

			err := repo.Store(*ranking)
			assert.NoError(t, err)
		}
	}

	// Create unrelated rankings to check if it is really working
	{
		for _, data := range []testCase{
			{contestID + 1, 1, domain.Global, 50},
			{contestID, 1, domain.Japanese, 250},
			{contestID, 2, domain.Korean, 150},
			{contestID + 1, 3, domain.Global, 200},
		} {
			ranking := &domain.Ranking{
				ContestID: data.contestID,
				UserID:    data.userID,
				Language:  data.language,
				Amount:    0,
			}

			err := repo.Store(*ranking)
			assert.NoError(t, err)
		}
	}

	rankings, err := repo.RankingsForContest(contestID, domain.Global)
	assert.NoError(t, err)

	assert.Equal(t, len(expected), len(rankings))

	for i, expected := range expected {
		// This assumption should work as the order of the rankings should be fixed
		ranking := rankings[i]

		assert.Equal(t, expected.amount, ranking.Amount)
		assert.Equal(t, contestID, ranking.ContestID)
		assert.Equal(t, expected.userID, ranking.UserID)
	}
}

func TestRankingRepository_GlobalRankings(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)
	userRepo := repositories.NewUserRepository(sqlHandler)

	contestID := uint64(1)

	expected := []struct {
		userID          uint64
		userDisplayName string
		language        domain.LanguageCode
		amount          float32
	}{
		{1, "FOO 1", domain.Global, 50},
		{3, "FOO 3", domain.Global, 30},
		{2, "FOO 2", domain.Global, 20},
	}

	{
		storedUsers := map[uint64]bool{}
		rankings := []struct {
			contestID uint64
			userID    uint64
			language  domain.LanguageCode
			amount    float32
		}{
			{contestID, 1, domain.Global, 40},
			{contestID + 1, 1, domain.Global, 10},
			{contestID + 1, 1, domain.Japanese, 10},
			{contestID, 2, domain.Global, 20},
			{contestID, 3, domain.Global, 30},
		}
		for _, data := range rankings {
			if storedUsers[data.userID] == false {
				storedUsers[data.userID] = true
				err := userRepo.Store(domain.User{
					Email:       fmt.Sprintf("foo+%d@bar.com", data.userID),
					DisplayName: fmt.Sprintf("FOO %d", data.userID),
					Password:    "foobar",
					Role:        domain.RoleUser,
					Preferences: &domain.Preferences{},
				})
				assert.NoError(t, err)
			}

			ranking := &domain.Ranking{
				ContestID: data.contestID,
				UserID:    data.userID,
				Language:  data.language,
				Amount:    data.amount,
			}

			err := repo.Store(*ranking)
			assert.NoError(t, err)
		}
	}

	rankings, err := repo.GlobalRankings(domain.Global)
	assert.NoError(t, err)

	assert.Equal(t, len(expected), len(rankings))

	for i, expected := range expected {
		// This assumption should work as the order of the rankings should be fixed
		ranking := rankings[i]

		assert.Equal(t, expected.amount, ranking.Amount)
		assert.Equal(t, expected.userID, ranking.UserID)
		assert.Equal(t, expected.userDisplayName, ranking.UserDisplayName)
	}
}
func TestRankingRepository_FindAllByContestAndUser(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)

	contestID := uint64(1)
	userID := uint64(1)

	expected := []struct {
		language domain.LanguageCode
		amount   float32
	}{
		{domain.Japanese, 10},
		{domain.Korean, 20},
		{domain.Global, 30},
	}

	// Correct rankings
	{
		for _, data := range expected {
			ranking := &domain.Ranking{
				ContestID: contestID,
				UserID:    userID,
				Language:  data.language,
				Amount:    data.amount,
			}

			err := repo.Store(*ranking)
			assert.NoError(t, err)
		}
	}

	// Create unrelated rankings to check if it is really working
	{
		for _, language := range []domain.LanguageCode{domain.Korean, domain.Global} {
			ranking := &domain.Ranking{
				ContestID: contestID,
				UserID:    userID + 1,
				Language:  language,
				Amount:    0,
			}

			err := repo.Store(*ranking)
			assert.NoError(t, err)
		}
	}

	rankings, err := repo.FindAll(contestID, userID)
	assert.NoError(t, err)

	for _, expected := range expected {
		var ranking domain.Ranking
		for _, r := range rankings {
			if r.Language == expected.language {
				ranking = r
			}
		}

		assert.Equal(t, expected.amount, ranking.Amount)
		assert.Equal(t, contestID, ranking.ContestID)
		assert.Equal(t, userID, ranking.UserID)
	}
}

func TestRankingRepository_UpdateAmounts(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)

	contestID := uint64(1)
	userID := uint64(1)

	// Create initial rankings
	for i, language := range []domain.LanguageCode{domain.Japanese, domain.Korean, domain.Global} {
		ranking := domain.Ranking{
			ContestID: contestID,
			UserID:    userID,
			Language:  language,
			Amount:    float32(i),
		}

		err := repo.Store(ranking)
		assert.NoError(t, err)
	}

	// Update rankings
	updatedRankings := domain.Rankings{}
	{
		rankings, err := repo.FindAll(contestID, userID)
		assert.NoError(t, err)

		for _, r := range rankings {
			updatedRankings = append(updatedRankings, domain.Ranking{
				ID:     r.ID,
				Amount: r.Amount + 10,
			})
		}

		err = repo.UpdateAmounts(updatedRankings)
		assert.NoError(t, err)
	}

	// Check updated content
	{
		rankings, err := repo.FindAll(contestID, userID)
		assert.NoError(t, err)

		assert.Equal(t, len(updatedRankings), len(rankings))

		expectedRankings := make(map[uint64]domain.Ranking)
		for _, ranking := range updatedRankings {
			expectedRankings[ranking.ID] = ranking
		}

		for _, ranking := range rankings {
			assert.Equal(t, ranking.Amount, expectedRankings[ranking.ID].Amount)
		}
	}
}
