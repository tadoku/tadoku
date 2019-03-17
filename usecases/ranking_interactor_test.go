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
	*usecases.MockUserRepository,
	*usecases.MockValidator,
	usecases.RankingInteractor,
) {
	ctrl := gomock.NewController(t)

	rankingRepo := usecases.NewMockRankingRepository(ctrl)
	contestRepo := usecases.NewMockContestRepository(ctrl)
	userRepo := usecases.NewMockUserRepository(ctrl)
	validator := usecases.NewMockValidator(ctrl)
	interactor := usecases.NewRankingInteractor(rankingRepo, contestRepo, userRepo, validator)

	return ctrl, rankingRepo, contestRepo, userRepo, validator, interactor
}

func TestRankingInteractor_CreateRanking(t *testing.T) {
	ctrl, rankingRepo, contestRepo, userRepo, _, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	{
		userID := uint64(1)
		contestID := uint64(1)
		languages := domain.LanguageCodes{domain.Japanese, domain.English}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)
		userRepo.EXPECT().FindByID(uint64(1)).Return(domain.User{ID: 1}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(uint64(1), uint64(1)).Return(nil, nil)
		rankingRepo.EXPECT().Store(gomock.Any()).Return(nil).Times(3) // Japanese, Chinese, and Global

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.NoError(t, err)
	}
}
