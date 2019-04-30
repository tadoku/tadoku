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
	*usecases.MockContestRepository,
	usecases.ContestLogInteractor,
) {
	ctrl := gomock.NewController(t)

	repo := usecases.NewMockContestLogRepository(ctrl)
	contestRepo := usecases.NewMockContestRepository(ctrl)
	interactor := usecases.NewContestLogInteractor(repo, contestRepo)

	return ctrl, repo, contestRepo, interactor
}

func TestContestLogInteractor_CreateLog(t *testing.T) {
	ctrl, repo, contestRepo, interactor := setupContestLogTest(t)
	defer ctrl.Finish()

	{
		log := domain.ContestLog{
			ContestID: 1,
			UserID:    1,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  1,
		}

		repo.EXPECT().Store(log)
		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)

		err := interactor.CreateLog(log)
		assert.NoError(t, err)
	}

	{
		log := domain.ContestLog{
			ContestID: 2,
			UserID:    1,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  1,
		}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)

		err := interactor.CreateLog(log)
		assert.Error(t, err)
	}
}
