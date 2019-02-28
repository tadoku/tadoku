package usecases_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"

	gomock "github.com/golang/mock/gomock"
)

func setupContestTest(t *testing.T) (
	*gomock.Controller,
	*usecases.MockContestRepository,
	usecases.ContestInteractor,
) {
	ctrl := gomock.NewController(t)

	repo := usecases.NewMockContestRepository(ctrl)
	interactor := usecases.NewContestInteractor(repo)

	return ctrl, repo, interactor
}

func TestSessionInteractor_CreateContest(t *testing.T) {
	ctrl, repo, interactor := setupContestTest(t)
	defer ctrl.Finish()

	contest := domain.Contest{
		Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
		Open:  true,
	}

	repo.EXPECT().Store(contest)

	err := interactor.CreateContest(contest)

	assert.NoError(t, err)
}
