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
	*usecases.MockValidator,
	usecases.RankingInteractor,
) {
	ctrl := gomock.NewController(t)

	repo := usecases.NewMockRankingRepository(ctrl)
	validator := usecases.NewMockValidator(ctrl)
	interactor := usecases.NewRankingInteractor(repo, validator)

	return ctrl, repo, validator, interactor
}

func TestRankingInteractor_CreateRanking(t *testing.T) {
	ctrl, repo, validator, interactor := setupRankingTest(t)
	defer ctrl.Finish()

	{
		repo.EXPECT().Store(ranking)
		repo.EXPECT().GetOpenRankings().Return(nil, nil)
		validator.EXPECT().Validate(ranking).Return(true, nil)

		userID := 1
		contestID := 1
		languages := domain.LanguageCodes{domain.Japanese, domain.English}

		err := interactor.CreateRanking(userID, contestID, languages)

		assert.NoError(t, err)
	}
}
