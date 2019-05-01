package usecases_test

import (
	"testing"

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
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.English, domain.Global}, nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: languages[0], Amount: 0}).Return(nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.NoError(t, err)
	}

	{
		languages := domain.LanguageCodes{domain.English}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)
		userRepo.EXPECT().FindByID(userID).Return(domain.User{ID: userID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.English, domain.Global}, nil)

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

		contestLogRepo.EXPECT().Store(log)
		validator.EXPECT().Validate(log).Return(true, nil)
		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Japanese}, nil)
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(rankings, nil)
		contestLogRepo.EXPECT().FindAll(contestID, userID).Return(domain.ContestLogs{log}, nil)
		rankingRepo.EXPECT().UpdateAmounts(expectedRankings).Return(nil)

		err := interactor.CreateLog(log)
		assert.NoError(t, err)
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
		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)

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
		contestRepo.EXPECT().GetOpenContests().Return([]uint64{contestID}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(contestID, userID).Return(domain.LanguageCodes{domain.Korean}, nil)

		err := interactor.CreateLog(log)
		assert.EqualError(t, err, usecases.ErrContestLanguageNotSignedUp.Error())
	}
}

func TestRankingInteractor_UpdateRankings(t *testing.T) {
	ctrl, rankingRepo, _, contestLogRepo, _, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	userID := uint64(1)

	{
		log := domain.ContestLog{
			ContestID: contestID,
			UserID:    userID,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  domain.MediumBook,
		}
		rankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 0},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 0},
		}
		expectedRankings := domain.Rankings{
			{ID: 1, ContestID: contestID, UserID: userID, Language: domain.Japanese, Amount: 10},
			{ID: 2, ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 10},
		}
		rankingRepo.EXPECT().FindAll(contestID, userID).Return(rankings, nil)
		contestLogRepo.EXPECT().FindAll(contestID, userID).Return(domain.ContestLogs{log}, nil)
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
