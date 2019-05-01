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

	{
		userID := uint64(1)
		contestID := uint64(1)
		languages := domain.LanguageCodes{domain.Japanese, domain.English}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)
		userRepo.EXPECT().FindByID(uint64(1)).Return(domain.User{ID: 1}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(uint64(1), uint64(1)).Return(nil, nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: languages[0], Amount: 0}).Return(nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: languages[1], Amount: 0}).Return(nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: domain.Global, Amount: 0}).Return(nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.NoError(t, err)
	}

	{
		userID := uint64(1)
		contestID := uint64(1)
		languages := domain.LanguageCodes{domain.Chinese}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)
		userRepo.EXPECT().FindByID(uint64(1)).Return(domain.User{ID: 1}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(uint64(1), uint64(1)).Return(domain.LanguageCodes{domain.English, domain.Global}, nil)
		rankingRepo.EXPECT().Store(domain.Ranking{ContestID: contestID, UserID: userID, Language: languages[0], Amount: 0}).Return(nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.NoError(t, err)
	}

	{
		userID := uint64(1)
		contestID := uint64(1)
		languages := domain.LanguageCodes{domain.English}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)
		userRepo.EXPECT().FindByID(uint64(1)).Return(domain.User{ID: 1}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(uint64(1), uint64(1)).Return(domain.LanguageCodes{domain.English, domain.Global}, nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.EqualError(t, err, usecases.ErrNoRankingToCreate.Error())
	}

	{
		userID := uint64(1)
		contestID := uint64(1)
		languages := domain.LanguageCodes{domain.Global}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)
		userRepo.EXPECT().FindByID(uint64(1)).Return(domain.User{ID: 1}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(uint64(1), uint64(1)).Return(nil, nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.EqualError(t, err, usecases.ErrGlobalIsASystemLanguage.Error())
	}

	{
		userID := uint64(1)
		contestID := uint64(1)
		languages := domain.LanguageCodes{"xxx"}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)
		userRepo.EXPECT().FindByID(uint64(1)).Return(domain.User{ID: 1}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(uint64(1), uint64(1)).Return(nil, nil)

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.EqualError(t, err, domain.ErrInvalidLanguage.Error())
	}
}
