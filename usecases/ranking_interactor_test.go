package usecases_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"

	gomock "github.com/golang/mock/gomock"
)

func setupRankingTest(t *testing.T) (
	*gomock.Controller,
	*usecases.MockRankingRepository,
	*usecases.MockContestRepository,
	*usecases.MockContestLogRepository,
	*usecases.MockUserRepository,
	*usecases.MockValidator,
	usecases.RankingInteractor,
) {
	ctrl := gomock.NewController(t)

	rankingRepo := usecases.NewMockRankingRepository(ctrl)
	contestRepo := usecases.NewMockContestRepository(ctrl)
	contestLogRepo := usecases.NewMockContestLogRepository(ctrl)
	userRepo := usecases.NewMockUserRepository(ctrl)
	validator := usecases.NewMockValidator(ctrl)
	interactor := usecases.NewRankingInteractor(rankingRepo, contestRepo, contestLogRepo, userRepo, validator)

	return ctrl, rankingRepo, contestRepo, contestLogRepo, userRepo, validator, interactor
}

func TestRankingInteractor_CreateRanking(t *testing.T) {
	ctrl, rankingRepo, contestRepo, _, userRepo, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	userID := uint64(1)

	{
		languages := domain.LanguageCodes{domain.Japanese, domain.English}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)
		userRepo.EXPECT().FindByID(userID).Return(domain.User{ID: userID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(nil, nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: languages[0], Amount: 0}).Return(nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: languages[1], Amount: 0}).Return(nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 0}).Return(nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.NoError(t, err)
	}

	{
		languages := domain.LanguageCodes{domain.Chinese}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)
		userRepo.EXPECT().FindByID(userID).Return(domain.User{ID: userID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.English}, nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: languages[0], Amount: 0}).Return(nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.NoError(t, err)
	}

	{
		languages := domain.LanguageCodes{domain.English}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)
		userRepo.EXPECT().FindByID(userID).Return(domain.User{ID: userID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.English}, nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.EqualError(t, err, usecases.ErrNoRankingToCreate.Error())
	}

	{
		languages := domain.LanguageCodes{domain.Global}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)
		userRepo.EXPECT().FindByID(userID).Return(domain.User{ID: userID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(nil, nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.EqualError(t, err, usecases.ErrGlobalIsASystemLanguage.Error())
	}

	{
		languages := domain.LanguageCodes{"xxx"}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)
		userRepo.EXPECT().FindByID(userID).Return(domain.User{ID: userID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(nil, nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.EqualError(t, err, domain.ErrInvalidLanguage.Error())
	}
}

func TestRankingInteractor_CreateLog(t *testing.T) {
	ctrl, rankingRepo, contestRepo, contestLogRepo, _, validator, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	userID := uint64(1)

	// Test happy path
	{
		log := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  domain.MediumManga,
		}

		rankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 0},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 0},
		}

		expectedRankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 2},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 2},
		}

		contestLogRepo.EXPECT().Store(&log)
		validator.EXPECT().Validate(log).Return(true, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Japanese}, nil)
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(rankings, nil)
		contestLogRepo.EXPECT().FindAll(contestID, userID).Return(domain.ContestLogs{log}, nil)
		rankingRepo.EXPECT().UpdateAmounts(expectedRankings).Return(nil)

		err := interactor.CreateLog(log)
		assert.NoError(t, err)
	}

	// Test creation with id
	{
		log := domain.ContestLog{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  20,
		}

		err := interactor.CreateLog(log)
		assert.EqualError(t, err, usecases.ErrCreateContestLogHasID.Error())
	}

	// Test invalid medium
	{
		log := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  20,
		}

		validator.EXPECT().Validate(log).Return(false, domain.ErrMediumNotFound)

		err := interactor.CreateLog(log)
		assert.EqualError(t, err, usecases.ErrInvalidContestLog.Error())
	}

	// Test contest being closed
	{
		log := domain.ContestLog{
			ContestID: contestID + 1,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  domain.MediumManga,
		}

		validator.EXPECT().Validate(log).Return(true, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)

		err := interactor.CreateLog(log)
		assert.EqualError(t, err, usecases.ErrContestIsClosed.Error())
	}

	// Test not being signed up for a language
	{
		log := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  domain.MediumManga,
		}

		validator.EXPECT().Validate(log).Return(true, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Korean}, nil)

		err := interactor.CreateLog(log)
		assert.EqualError(t, err, usecases.ErrContestLanguageNotSignedUp.Error())
	}

	// Test global is not accepted as a language
	{
		log := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Global,
			Amount:    10,
			MediumID:  domain.MediumManga,
		}

		validator.EXPECT().Validate(log).Return(true, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Japanese}, nil)

		err := interactor.CreateLog(log)
		assert.EqualError(t, err, usecases.ErrContestLanguageNotSignedUp.Error())
	}
}

func TestRankingInteractor_UpdateLog(t *testing.T) {
	ctrl, rankingRepo, contestRepo, contestLogRepo, _, validator, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	userID := uint64(1)

	{
		log := domain.ContestLog{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  domain.MediumManga,
		}

		rankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 0},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 0},
		}

		expectedRankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 2},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 2},
		}

		contestLogRepo.EXPECT().Store(&log)
		validator.EXPECT().Validate(log).Return(true, nil)
		contestLogRepo.EXPECT().FindByID(log.ID).Return(log, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Japanese}, nil)
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(rankings, nil)
		contestLogRepo.EXPECT().FindAll(contestID, userID).Return(domain.ContestLogs{log}, nil)
		rankingRepo.EXPECT().UpdateAmounts(expectedRankings).Return(nil)

		err := interactor.UpdateLog(log)

		assert.NoError(t, err)
	}

	{
		log := domain.ContestLog{
			ID:        1,
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  domain.MediumManga,
		}

		validator.EXPECT().Validate(log).Return(true, nil)
		contestLogRepo.EXPECT().FindByID(log.ID).Return(domain.ContestLog{UserID: userID + 1}, nil)

		err := interactor.UpdateLog(log)

		assert.EqualError(t, err, domain.ErrInsufficientPermissions.Error())
	}

	{
		log := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Global,
			Amount:    10,
			MediumID:  domain.MediumManga,
		}

		err := interactor.UpdateLog(log)

		assert.EqualError(t, err, usecases.ErrContestLogIDMissing.Error())
	}
}

func TestRankingInteractor_DeleteLog(t *testing.T) {
	ctrl, rankingRepo, contestRepo, contestLogRepo, _, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	userID := uint64(1)

	log := domain.ContestLog{
		ID:        1,
		ContestID: contestID,
		UserID:    userID,
		Language:  domain.Japanese,
		Amount:    10,
		MediumID:  domain.MediumBook,
	}

	// Happy path
	{
		rankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 10},
		}

		expectedRankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 0},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 0},
		}

		contestLogRepo.EXPECT().Delete(log.ID)
		contestLogRepo.EXPECT().FindByID(log.ID).Return(log, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(rankings, nil)
		contestLogRepo.EXPECT().FindAll(contestID, userID).Return(domain.ContestLogs{}, nil)
		rankingRepo.EXPECT().UpdateAmounts(expectedRankings).Return(nil)

		err := interactor.DeleteLog(log.ID, log.UserID)
		assert.NoError(t, err)
	}

	// Sad path: different user trying to delete a log
	{
		contestLogRepo.EXPECT().FindByID(log.ID).Return(log, nil)

		err := interactor.DeleteLog(log.ID, log.UserID+1)
		assert.EqualError(t, err, domain.ErrInsufficientPermissions.Error())
	}

	// Sad path: contest is cloaed
	{
		contestLogRepo.EXPECT().FindByID(log.ID).Return(log, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID + 1}, nil)

		err := interactor.DeleteLog(log.ID, log.UserID)
		assert.EqualError(t, err, usecases.ErrContestIsClosed.Error())
	}

	// Sad path: log does not exist
	{
		contestLogRepo.EXPECT().FindByID(log.ID).Return(domain.ContestLog{}, domain.ErrNotFound)

		err := interactor.DeleteLog(log.ID, log.UserID)
		assert.EqualError(t, err, domain.ErrNotFound.Error())
	}
}

func TestRankingInteractor_UpdateRankings(t *testing.T) {
	ctrl, rankingRepo, _, contestLogRepo, _, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	userID := uint64(1)

	{
		logJapaneseBook := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  domain.MediumBook,
		}
		logKoreanManga := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Korean,
			Amount:    10,
			MediumID:  domain.MediumManga,
		}
		rankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 0},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Korean, Amount: 0},
			{ID: 3, ContestID: contestID, UserID: userID, Language: domain.German, Amount: 11},
			{ID: 4, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 0},
		}
		expectedRankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Korean, Amount: 2},
			{ID: 3, ContestID: contestID, UserID: userID, Language: domain.German, Amount: 0},
			{ID: 4, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 12},
		}
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(rankings, nil)
		contestLogRepo.EXPECT().FindAll(contestID, userID).Return(domain.ContestLogs{logJapaneseBook, logKoreanManga}, nil)
		rankingRepo.EXPECT().UpdateAmounts(expectedRankings).Return(nil)

		err := interactor.UpdateRanking(contestID, userID)
		assert.NoError(t, err)
	}

	{
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(nil, nil)

		err := interactor.UpdateRanking(contestID, userID)
		assert.EqualError(t, err, usecases.ErrNoRankingsFound.Error())
	}
}

func TestRankingInteractor_RankingsForRegistration(t *testing.T) {
	ctrl, rankingRepo, _, _, _, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	userID := uint64(1)

	{
		expected := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Korean, Amount: 2},
			{ID: 3, ContestID: contestID, UserID: userID, Language: domain.German, Amount: 0},
			{ID: 4, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 12},
		}
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(expected, nil)

		rankings, err := interactor.RankingsForRegistration(contestID, userID)
		assert.NoError(t, err)
		assert.Equal(t, expected, rankings)
	}

	{
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(nil, nil)

		rankings, err := interactor.RankingsForRegistration(contestID, userID)
		assert.EqualError(t, err, usecases.ErrNoRankingsFound.Error())
		assert.Equal(t, 0, len(rankings))
	}
}

func TestRankingInteractor_RankingsForContest(t *testing.T) {
	ctrl, rankingRepo, _, _, _, validator, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	userID := uint64(1)
	language := domain.Global

	// Happy path for specific contest
	{
		expected := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: language, Amount: 15},
			{ID: 2, ContestID: contestID, UserID: userID + 1, Language: language, Amount: 12},
			{ID: 3, ContestID: contestID, UserID: userID + 2, Language: language, Amount: 11},
			{ID: 4, ContestID: contestID, UserID: userID + 3, Language: language, Amount: 0},
		}
		rankingRepo.EXPECT().RankingsForContest(contestID, language).Return(expected, nil)
		validator.EXPECT().Validate(language).Return(true, nil)

		rankings, err := interactor.RankingsForContest(contestID, language)
		assert.NoError(t, err)

		for i, ranking := range rankings {
			expect := expected[i]

			assert.Equal(t, expect.ID, ranking.ID)
			assert.Equal(t, expect.Amount, ranking.Amount)
		}
	}

	// Happy path for global rankings
	{
		expected := domain.Rankings{
			{UserID: userID, Language: language, Amount: 15},
			{UserID: userID + 1, Language: language, Amount: 12},
			{UserID: userID + 2, Language: language, Amount: 11},
			{UserID: userID + 3, Language: language, Amount: 0},
		}
		rankingRepo.EXPECT().GlobalRankings(language).Return(expected, nil)
		validator.EXPECT().Validate(language).Return(true, nil)

		rankings, err := interactor.RankingsForContest(0, language)
		assert.NoError(t, err)

		for i, ranking := range rankings {
			expect := expected[i]

			assert.Equal(t, expect.UserID, ranking.UserID)
			assert.Equal(t, expect.Amount, ranking.Amount)
		}
	}

	// Happy path for specific language and contest
	{
		expected := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 15},
		}
		rankingRepo.EXPECT().RankingsForContest(contestID, domain.Japanese).Return(expected, nil)
		validator.EXPECT().Validate(domain.Japanese).Return(true, nil)

		_, err := interactor.RankingsForContest(contestID, domain.Japanese)
		assert.NoError(t, err)
	}

	// Happy path for invalid language falls back to global
	{
		invalidLanguage := domain.LanguageCode("")
		expected := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: language, Amount: 15},
		}
		rankingRepo.EXPECT().RankingsForContest(contestID, language).Return(expected, nil)
		validator.EXPECT().Validate(invalidLanguage).Return(false, domain.ErrInvalidLanguage)

		_, err := interactor.RankingsForContest(contestID, invalidLanguage)
		assert.NoError(t, err)
	}

	// Sad path for no rankings found
	{
		rankingRepo.EXPECT().RankingsForContest(contestID, language).Return(nil, nil)
		validator.EXPECT().Validate(language).Return(true, nil)

		_, err := interactor.RankingsForContest(contestID, language)
		assert.EqualError(t, err, usecases.ErrNoRankingsFound.Error())
	}
}

func TestRankingInteractor_CurrentRegistration(t *testing.T) {
	ctrl, rankingRepo, _, _, _, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	userID := uint64(1)

	{
		// Happy path
		expected := domain.RankingRegistration{
			ContestID: 1,
			End:       time.Now(),
			Languages: domain.LanguageCodes{domain.Japanese, domain.Korean},
		}
		rankingRepo.EXPECT().CurrentRegistration(userID).Return(expected, nil)

		registration, err := interactor.CurrentRegistration(userID)
		assert.NoError(t, err)
		assert.Equal(t, expected, registration)
	}

	{
		// Sad path no registration found
		rankingRepo.EXPECT().CurrentRegistration(userID).Return(domain.RankingRegistration{}, nil)

		_, err := interactor.CurrentRegistration(userID)
		assert.EqualError(t, err, usecases.ErrNoRankingRegistrationFound.Error())
	}

}

func TestRankingInteractor_ContestLogs(t *testing.T) {
	ctrl, _, _, repo, _, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	userID := uint64(1)
	contestID := uint64(1)

	{
		// Happy path
		expected := domain.ContestLogs{
			{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumBook, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 20, MediumID: domain.MediumManga, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: userID, Language: domain.Chinese, Amount: 30, MediumID: domain.MediumGame, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: userID, Language: domain.Korean, Amount: 100, MediumID: domain.MediumNet, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 40, MediumID: domain.MediumBook, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
		}
		repo.EXPECT().FindAll(contestID, userID).Return(expected, nil)

		logs, err := interactor.ContestLogs(contestID, userID)
		assert.NoError(t, err)
		assert.Equal(t, expected, logs)
	}

	{
		// Sad path no registration found
		repo.EXPECT().FindAll(contestID, userID).Return(domain.ContestLogs{}, nil)

		_, err := interactor.ContestLogs(contestID, userID)
		assert.EqualError(t, err, usecases.ErrNoContestLogsFound.Error())
	}
}
