package repositories_test

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/repositories"
)

func TestRankingRepository_StoreRanking(t *testing.T) {
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
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)

	contestID := uint64(1)
	users := createTestUsers(t, sqlHandler, 3)

	type testCase struct {
		contestID       uint64
		userID          uint64
		userDisplayName string
		language        domain.LanguageCode
		amount          float32
	}
	expected := []testCase{
		{contestID, users[2].ID, "FOO 3", domain.Global, 30},
		{contestID, users[1].ID, "FOO 2", domain.Global, 20},
		{contestID, users[0].ID, "FOO 1", domain.Global, 10},
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
			{contestID + 1, 1, "", domain.Global, 50},
			{contestID, 1, "", domain.Japanese, 250},
			{contestID, 2, "", domain.Korean, 150},
			{contestID + 1, 3, "", domain.Global, 200},
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
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)

	contestID := uint64(1)
	users := createTestUsers(t, sqlHandler, 3)

	expected := []struct {
		userID          uint64
		userDisplayName string
		language        domain.LanguageCode
		amount          float32
	}{
		{users[0].ID, users[0].DisplayName, domain.Global, 50},
		{users[2].ID, users[2].DisplayName, domain.Global, 30},
		{users[1].ID, users[1].DisplayName, domain.Global, 20},
	}

	{
		rankings := []struct {
			contestID uint64
			userID    uint64
			language  domain.LanguageCode
			amount    float32
		}{
			{contestID, users[0].ID, domain.Global, 40},
			{contestID + 1, users[0].ID, domain.Global, 10},
			{contestID + 1, users[0].ID, domain.Japanese, 10},
			{contestID, users[1].ID, domain.Global, 20},
			{contestID, users[2].ID, domain.Global, 30},
		}
		for _, data := range rankings {
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
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)

	contestID := uint64(1)
	users := createTestUsers(t, sqlHandler, 2)

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
				UserID:    users[0].ID,
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
				UserID:    users[1].ID,
				Language:  language,
				Amount:    0,
			}

			err := repo.Store(*ranking)
			assert.NoError(t, err)
		}
	}

	rankings, err := repo.FindAll(contestID, users[0].ID)
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
		assert.Equal(t, users[0].ID, ranking.UserID)
		assert.Equal(t, users[0].DisplayName, ranking.UserDisplayName)
	}

	{
		rankings, err := repo.FindAll(0, 0)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(rankings))
	}
}

func TestRankingRepository_UpdateAmounts(t *testing.T) {
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

func TestRankingRepository_CurrentRegistration(t *testing.T) {
	sqlHandler, cleanup := setupTestingSuite(t)
	defer cleanup()

	repo := repositories.NewRankingRepository(sqlHandler)
	contestRepo := repositories.NewContestRepository(sqlHandler)

	user := createTestUsers(t, sqlHandler, 1)[0]
	languages := domain.LanguageCodes{domain.Japanese, domain.German}
	contest := &domain.Contest{
		Description: "Round foo",
		Start:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		End:         time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
		Open:        true,
	}

	now := time.Date(2019, 1, 20, 0, 0, 0, 0, time.UTC)

	{
		err := contestRepo.Store(contest)
		assert.NoError(t, err)
	}

	{
		for _, l := range languages {
			err := repo.Store(domain.Ranking{
				ContestID: contest.ID,
				UserID:    user.ID,
				Language:  l,
				Amount:    0,
			})
			assert.NoError(t, err)
		}
	}

	{
		registration, err := repo.CurrentRegistration(user.ID, now)
		assert.NoError(t, err)
		assert.Equal(t, contest.ID, registration.ContestID)
		assert.Equal(t, contest.Start.UTC(), registration.Start.UTC())
		assert.Equal(t, contest.End.UTC(), registration.End.UTC())

		sort.Sort(languages)
		sort.Sort(registration.Languages)
		assert.Equal(t, languages, registration.Languages)
	}
}
