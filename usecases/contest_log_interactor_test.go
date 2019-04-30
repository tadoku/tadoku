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
	*usecases.MockRankingRepository,
	*usecases.MockValidator,
	usecases.ContestLogInteractor,
) {
	ctrl := gomock.NewController(t)

	repo := usecases.NewMockContestLogRepository(ctrl)
	contestRepo := usecases.NewMockContestRepository(ctrl)
	rankingRepo := usecases.NewMockRankingRepository(ctrl)
	validator := usecases.NewMockValidator(ctrl)
	interactor := usecases.NewContestLogInteractor(repo, contestRepo, rankingRepo, validator)

	return ctrl, repo, contestRepo, rankingRepo, validator, interactor
}

func TestContestLogInteractor_CreateLog(t *testing.T) {
	ctrl, repo, contestRepo, rankingRepo, _, interactor := setupContestLogTest(t)
	defer ctrl.Finish()

	// Test happy path
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
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(uint64(1), uint64(1)).Return(domain.LanguageCodes{domain.Japanese}, nil)

		err := interactor.CreateLog(log)
		assert.NoError(t, err)
	}

	// Test contest being closed
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
		assert.Equal(t, err, usecases.ErrContestIsClosed)
	}

	// Test not being signed up for a language
	{
		log := domain.ContestLog{
			ContestID: 1,
			UserID:    1,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  1,
		}

		contestRepo.EXPECT().GetOpenContests().Return([]uint64{1}, nil)
		rankingRepo.EXPECT().GetAllLanguagesForContestAndUser(uint64(1), uint64(1)).Return(domain.LanguageCodes{domain.Korean}, nil)

		err := interactor.CreateLog(log)
		assert.Equal(t, err, usecases.ErrContestLanguageNotSignedUp)
	}
}
