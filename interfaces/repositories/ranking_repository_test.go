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

func TestRankingRepository_UpdateRankingsForContestAndUser(t *testing.T) {
	t.Parallel()
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)
	logsRepo := repositories.NewContestLogRepository(sqlHandler)

	contestID := uint64(1)
	userID := uint64(1)

	for _, language := range []domain.LanguageCode{domain.Japanese, domain.Korean, domain.Global} {
		ranking := &domain.Ranking{
			ContestID: contestID,
			UserID:    userID,
			Language:  language,
			Amount:    0,
			CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		err := repo.Store(*ranking)
		assert.NoError(t, err)
	}

	logs := []domain.ContestLog{
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumBook},      // 10 pages
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumManga},     // 2 pages
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumNet},       // 10 pages
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumFullGame},  // 1.667 pages
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Korean, Amount: 10, MediumID: domain.MediumGame},        // 0.5 pages
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumLyric},     // 10 pages
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumSubs},      // 2 pages
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Korean, Amount: 10, MediumID: domain.MediumNews},        // 10 pages
		domain.ContestLog{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumSentences}, // 0.5 pages
	}

	for _, log := range logs {
		err := logsRepo.Store(log)
		assert.NoError(t, err)
	}

	{
		err := repo.UpdateRankingsForContestAndUser(contestID, userID)
		assert.NoError(t, err)
	}

	// TODO: check actual rankings result
}
