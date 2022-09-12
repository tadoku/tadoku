package usecases_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/tadoku/services/reading-contest-api/domain"
	"github.com/tadoku/tadoku/services/reading-contest-api/usecases"

	gomock "github.com/golang/mock/gomock"
)

func setupRankingTest(t *testing.T) (
	*gomock.Controller,
	*usecases.MockRankingRepository,
	*usecases.MockContestRepository,
	*usecases.MockContestLogRepository,
	*usecases.MockValidator,
	usecases.RankingInteractor,
) {
	ctrl := gomock.NewController(t)

	rankingRepo := usecases.NewMockRankingRepository(ctrl)
	contestRepo := usecases.NewMockContestRepository(ctrl)
	contestLogRepo := usecases.NewMockContestLogRepository(ctrl)
	validator := usecases.NewMockValidator(ctrl)
	interactor := usecases.NewRankingInteractor(rankingRepo, contestRepo, contestLogRepo, validator)

	return ctrl, rankingRepo, contestRepo, contestLogRepo, validator, interactor
}

func TestRankingInteractor_CreateLog(t *testing.T) {
	ctrl, rankingRepo, contestRepo, contestLogRepo, validator, interactor := setupRankingTest(t)
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
			MediumID:  domain.MediumComic,
		}

		contestLogRepo.EXPECT().Store(&log)
		validator.EXPECT().Validate(log).Return(true, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Japanese}, nil)

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

	// Test not being signed up for a language
	{
		log := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  domain.MediumComic,
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
			MediumID:  domain.MediumComic,
		}

		validator.EXPECT().Validate(log).Return(true, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Japanese}, nil)

		err := interactor.CreateLog(log)
		assert.EqualError(t, err, usecases.ErrContestLanguageNotSignedUp.Error())
	}
}

func TestRankingInteractor_UpdateLog(t *testing.T) {
	ctrl, rankingRepo, contestRepo, contestLogRepo, validator, interactor := setupRankingTest(t)
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
			MediumID:  domain.MediumComic,
		}

		contestLogRepo.EXPECT().Store(&log)
		validator.EXPECT().Validate(log).Return(true, nil)
		contestLogRepo.EXPECT().FindByID(log.ID).Return(log, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Japanese}, nil)

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
			MediumID:  domain.MediumComic,
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
			MediumID:  domain.MediumComic,
		}

		err := interactor.UpdateLog(log)

		assert.EqualError(t, err, usecases.ErrContestLogIDMissing.Error())
	}
}

func TestRankingInteractor_DeleteLog(t *testing.T) {
	ctrl, _, contestRepo, contestLogRepo, _, interactor := setupRankingTest(t)
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

		contestLogRepo.EXPECT().Delete(log.ID)
		contestLogRepo.EXPECT().FindByID(log.ID).Return(log, nil)
		contestRepo.EXPECT().GetRunningContests().Return([]uint64{contestID}, nil)

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

func TestRankingInteractor_RankingsForRegistration(t *testing.T) {
	ctrl, rankingRepo, _, _, _, interactor := setupRankingTest(t)
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
	ctrl, rankingRepo, _, _, _, interactor := setupRankingTest(t)
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
		rankingRepo.EXPECT().RankingsForContest(contestID).Return(expected, nil)

		rankings, err := interactor.RankingsForContest(contestID)
		assert.NoError(t, err)

		for i, ranking := range rankings {
			expect := expected[i]

			assert.Equal(t, expect.ID, ranking.ID)
			assert.Equal(t, expect.Amount, ranking.Amount)
		}
	}

	// Sad path for no rankings found
	{
		rankingRepo.EXPECT().RankingsForContest(contestID).Return(nil, nil)

		_, err := interactor.RankingsForContest(contestID)
		assert.EqualError(t, err, usecases.ErrNoRankingsFound.Error())
	}
}

func TestRankingInteractor_CurrentRegistration(t *testing.T) {
	ctrl, rankingRepo, _, _, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	userID := uint64(1)

	{
		// Happy path
		expected := domain.RankingRegistration{
			ContestID: 1,
			Start:     time.Now().Add(-10 * time.Minute),
			End:       time.Now().Add(10 * time.Minute),
			Languages: domain.LanguageCodes{domain.Japanese, domain.Korean},
		}
		rankingRepo.EXPECT().CurrentRegistration(userID, gomock.Any()).Return(expected, nil)

		registration, err := interactor.CurrentRegistration(userID)
		assert.NoError(t, err)
		assert.Equal(t, expected, registration)
	}

	{
		// Sad path no registration found
		rankingRepo.EXPECT().CurrentRegistration(userID, gomock.Any()).Return(domain.RankingRegistration{}, nil)

		_, err := interactor.CurrentRegistration(userID)
		assert.EqualError(t, err, usecases.ErrNoRankingRegistrationFound.Error())
	}

}

func TestRankingInteractor_ContestLogs(t *testing.T) {
	ctrl, _, _, repo, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	userID := uint64(1)
	contestID := uint64(1)

	{
		// Happy path
		expected := domain.ContestLogs{
			{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumBook, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 20, MediumID: domain.MediumComic, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
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

func TestRankingInteractor_RecentContestLogs(t *testing.T) {
	ctrl, _, _, repo, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	limit := uint64(50)

	{
		// Happy path
		expected := domain.ContestLogs{
			{ContestID: contestID, UserID: 1, Language: domain.Japanese, Amount: 10, MediumID: domain.MediumBook, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: 1, Language: domain.Japanese, Amount: 20, MediumID: domain.MediumComic, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: 1, Language: domain.Chinese, Amount: 30, MediumID: domain.MediumGame, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: 2, Language: domain.Korean, Amount: 100, MediumID: domain.MediumNet, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			{ContestID: contestID, UserID: 2, Language: domain.Japanese, Amount: 40, MediumID: domain.MediumBook, CreatedAt: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
		}
		repo.EXPECT().FindRecent(contestID, limit).Return(expected, nil)

		logs, err := interactor.RecentContestLogs(contestID, limit)
		assert.NoError(t, err)
		assert.Equal(t, expected, logs)
	}

	{
		// Sad path no logs found for contest
		repo.EXPECT().FindRecent(contestID, limit).Return(domain.ContestLogs{}, nil)

		_, err := interactor.RecentContestLogs(contestID, limit)
		assert.EqualError(t, err, usecases.ErrNoContestLogsFound.Error())
	}

}
