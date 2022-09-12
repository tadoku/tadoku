package repositories_test

// TODO: Move tests which will become relevant again after migration
// func TestRankingRepository_GetAllLanguagesForContestAndUser(t *testing.T) {
// 	sqlHandler, cleanup := setupTestingSuite(t)
// 	defer cleanup()

// 	repo := repositories.NewRankingRepository(sqlHandler)
// 	rankingJapanese := &domain.Ranking{
// 		ContestID: 1,
// 		UserID:    1,
// 		Language:  domain.Japanese,
// 		Amount:    0,
// 		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	rankingChinese := &domain.Ranking{
// 		ContestID: 1,
// 		UserID:    1,
// 		Language:  domain.Chinese,
// 		Amount:    0,
// 		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	rankingGlobal := &domain.Ranking{
// 		ContestID: 1,
// 		UserID:    1,
// 		Language:  domain.Global,
// 		Amount:    0,
// 		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	rankingSingleLanguage := &domain.Ranking{
// 		ContestID: 1,
// 		UserID:    2,
// 		Language:  domain.Chinese,
// 		Amount:    0,
// 		CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 		UpdatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 	}

// 	{
// 		for _, r := range []*domain.Ranking{rankingJapanese, rankingChinese, rankingGlobal, rankingSingleLanguage} {
// 			err := repo.Store(*r)
// 			assert.NoError(t, err)
// 		}
// 	}

// 	{
// 		languages, err := repo.GetAllLanguagesForContestAndUser(1, 1)
// 		assert.NoError(t, err)
// 		assert.Equal(t, len(languages), 2)
// 		assert.Equal(t, languages[0], domain.Japanese)
// 		assert.Equal(t, languages[1], domain.Chinese)
// 	}

// 	{
// 		languages, err := repo.GetAllLanguagesForContestAndUser(1, 2)
// 		assert.NoError(t, err)
// 		assert.Equal(t, len(languages), 1)
// 		assert.Equal(t, languages[0], domain.Chinese)
// 	}
// }

// func TestRankingRepository_RankingsForContest(t *testing.T) {
// 	sqlHandler, cleanup := setupTestingSuite(t)
// 	defer cleanup()

// 	repo := repositories.NewRankingRepository(sqlHandler)
// 	logsRepo := repositories.NewContestLogRepository(sqlHandler)

// 	contestID := uint64(1)
// 	users := createTestUsers(t, sqlHandler, 3)

// 	type testCase struct {
// 		contestID uint64
// 		user      *domain.User
// 		language  domain.LanguageCode
// 		amount    float32
// 		logs      []domain.ContestLog
// 	}
// 	expected := []testCase{
// 		{contestID, users[2], domain.Global, 30, []domain.ContestLog{
// 			{ContestID: contestID, UserID: users[2].ID, Language: domain.German, MediumID: domain.MediumBook, Amount: 10},
// 			{ContestID: contestID, UserID: users[2].ID, Language: domain.German, MediumID: domain.MediumNet, Amount: 10},
// 			{ContestID: contestID, UserID: users[2].ID, Language: domain.German, MediumID: domain.MediumComic, Amount: 50},
// 		}},
// 		{contestID, users[1], domain.Global, 20, []domain.ContestLog{
// 			{ContestID: contestID, UserID: users[1].ID, Language: domain.German, MediumID: domain.MediumBook, Amount: 10},
// 			{ContestID: contestID, UserID: users[1].ID, Language: domain.German, MediumID: domain.MediumNet, Amount: 10},
// 		}},
// 		{contestID, users[0], domain.Global, 10, []domain.ContestLog{
// 			{ContestID: contestID, UserID: users[0].ID, Language: domain.German, MediumID: domain.MediumBook, Amount: 10},
// 		}},
// 	}

// 	// Correct rankings
// 	{
// 		for _, data := range []testCase{expected[2], expected[1], expected[0]} {
// 			ranking := &domain.Ranking{
// 				ContestID:       data.contestID,
// 				UserID:          data.user.ID,
// 				UserDisplayName: data.user.DisplayName,
// 				Language:        data.language,
// 				Amount:          data.amount,
// 			}

// 			err := repo.Store(*ranking)
// 			assert.NoError(t, err)

// 			for _, log := range data.logs {
// 				err := logsRepo.Store(&log)
// 				assert.NoError(t, err)
// 			}
// 		}
// 	}

// 	rankings, err := repo.RankingsForContest(contestID)
// 	assert.NoError(t, err)

// 	assert.Equal(t, len(expected), len(rankings))

// 	for i, expected := range expected {
// 		// This assumption should work as the order of the rankings should be fixed
// 		ranking := rankings[i]

// 		assert.Equal(t, expected.amount, ranking.Amount)
// 		assert.Equal(t, contestID, ranking.ContestID)
// 		assert.Equal(t, expected.user.ID, ranking.UserID)
// 		assert.Equal(t, expected.user.DisplayName, ranking.UserDisplayName)
// 	}
// }

// func TestRankingRepository_FindAllByContestAndUser(t *testing.T) {
// 	sqlHandler, cleanup := setupTestingSuite(t)
// 	defer cleanup()

// 	repo := repositories.NewRankingRepository(sqlHandler)
// 	contestRepo := repositories.NewContestLogRepository(sqlHandler)

// 	contestID := uint64(1)
// 	users := createTestUsers(t, sqlHandler, 2)

// 	expected := []struct {
// 		language domain.LanguageCode
// 		amount   float32
// 	}{
// 		{domain.Japanese, 10},
// 		{domain.Korean, 20},
// 	}

// 	// Correct rankings
// 	{
// 		for _, data := range expected {
// 			ranking := &domain.Ranking{
// 				ContestID:       contestID,
// 				UserID:          users[0].ID,
// 				Language:        data.language,
// 				UserDisplayName: users[0].DisplayName,
// 			}

// 			err := repo.Store(*ranking)
// 			assert.NoError(t, err)

// 			log := &domain.ContestLog{
// 				ContestID: contestID,
// 				UserID:    users[0].ID,
// 				Language:  data.language,
// 				Amount:    data.amount,
// 				MediumID:  domain.MediumBook,
// 			}

// 			err = contestRepo.Store(log)
// 			assert.NoError(t, err)
// 		}
// 	}

// 	// Create unrelated rankings to check if it is really working
// 	{
// 		for _, language := range []domain.LanguageCode{domain.Korean, domain.Global} {
// 			ranking := &domain.Ranking{
// 				ContestID:       contestID,
// 				UserID:          users[1].ID,
// 				Language:        language,
// 				UserDisplayName: users[1].DisplayName,
// 			}

// 			err := repo.Store(*ranking)
// 			assert.NoError(t, err)
// 		}
// 	}

// 	rankings, err := repo.FindAll(contestID, users[0].ID)
// 	assert.NoError(t, err)

// 	for _, expected := range expected {
// 		var ranking domain.Ranking
// 		for _, r := range rankings {
// 			if r.Language == expected.language {
// 				ranking = r
// 			}
// 		}

// 		assert.Equal(t, expected.amount, ranking.Amount)
// 		assert.Equal(t, contestID, ranking.ContestID)
// 		assert.Equal(t, users[0].ID, ranking.UserID)
// 		assert.Equal(t, users[0].DisplayName, ranking.UserDisplayName)
// 	}

// 	{
// 		rankings, err := repo.FindAll(0, 0)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 0, len(rankings))
// 	}
// }

// func TestRankingRepository_CurrentRegistration(t *testing.T) {
// 	sqlHandler, cleanup := setupTestingSuite(t)
// 	defer cleanup()

// 	repo := repositories.NewRankingRepository(sqlHandler)
// 	contestRepo := repositories.NewContestRepository(sqlHandler)

// 	user := createTestUsers(t, sqlHandler, 1)[0]
// 	languages := domain.LanguageCodes{domain.Japanese, domain.German}
// 	contest := &domain.Contest{
// 		Description: "Round foo",
// 		Start:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
// 		End:         time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
// 		Open:        true,
// 	}

// 	now := time.Date(2019, 1, 20, 0, 0, 0, 0, time.UTC)

// 	{
// 		err := contestRepo.Store(contest)
// 		assert.NoError(t, err)
// 	}

// 	{
// 		for _, l := range languages {
// 			err := repo.Store(domain.Ranking{
// 				ContestID: contest.ID,
// 				UserID:    user.ID,
// 				Language:  l,
// 				Amount:    0,
// 			})
// 			assert.NoError(t, err)
// 		}
// 	}

// 	{
// 		registration, err := repo.CurrentRegistration(user.ID, now)
// 		assert.NoError(t, err)
// 		assert.Equal(t, contest.ID, registration.ContestID)
// 		assert.Equal(t, contest.Start.UTC(), registration.Start.UTC())
// 		assert.Equal(t, contest.End.UTC(), registration.End.UTC())

// 		sort.Sort(languages)
// 		sort.Sort(registration.Languages)
// 		assert.Equal(t, languages, registration.Languages)
// 	}
// }
