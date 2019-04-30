package usecases_test

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"
)

func setupContestLogTest(t *testing.T) (
	*gomock.Controller,
	*usecases.MockContestLogRepository,
	usecases.ContestLogInteractor,
) {
	ctrl := gomock.NewController(t)

	repo := usecases.NewMockContestLogRepository(ctrl)
	interactor := usecases.NewContestLogInteractor(repo)

	return ctrl, repo, interactor
}

func TestContestLogInteractor_CreateLog(t *testing.T) {
	ctrl, repo, interactor := setupContestLogTest(t)
	defer ctrl.Finish()

	{
		log := domain.ContestLog{}

		repo.EXPECT().Store(log)

		err := interactor.CreateLog(log)

		assert.NoError(t, err)
	}
}
