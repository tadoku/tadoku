package repositories_test

import (
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
		assert.Equal(t, len(languages), 3)
		assert.Equal(t, languages[0], domain.Japanese)
		assert.Equal(t, languages[1], domain.Chinese)
		assert.Equal(t, languages[2], domain.Global)
	}

	{
		languages, err := repo.GetAllLanguagesForContestAndUser(1, 2)
		assert.NoError(t, err)
		assert.Equal(t, len(languages), 1)
		assert.Equal(t, languages[0], domain.Chinese)
	}
}

func TestRankingRepository_FindAllByContestAndUser(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)

	contestID := uint64(1)
	userID := uint64(1)

	// Correct rankings
	originalRankings := []domain.Ranking{}
	for i, language := range []domain.LanguageCode{domain.Japanese, domain.Korean, domain.Global} {
		ranking := &domain.Ranking{
			ContestID: contestID,
			UserID:    userID,
			Language:  language,
			Amount:    float32(i),
			CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		originalRankings = append(originalRankings, *ranking)

		err := repo.Store(*ranking)
		assert.NoError(t, err)
	}

	// Unrelated rankings
	for i, language := range []domain.LanguageCode{domain.Korean, domain.Global} {
		ranking := &domain.Ranking{
			ContestID: contestID,
			UserID:    userID + 1,
			Language:  language,
			Amount:    float32(i),
			CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		err := repo.Store(*ranking)
		assert.NoError(t, err)
	}

	rankings, err := repo.FindAll(contestID, userID)
	assert.NoError(t, err)

	assert.Equal(t, len(originalRankings), len(rankings))

	for i, ranking := range originalRankings {
		r := rankings[i]
		assert.Equal(t, uint64(i+1), r.ID)
		assert.Equal(t, ranking.ContestID, r.ContestID)
		assert.Equal(t, ranking.UserID, r.UserID)
		assert.Equal(t, ranking.Language, r.Language)
		assert.Equal(t, ranking.Amount, r.Amount)
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
