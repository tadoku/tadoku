package services_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/services"
	"github.com/tadoku/api/usecases"
)

func TestContestService_Create(t *testing.T) {
	contest := &domain.Contest{
		Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
		Open:  true,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(201)
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *contest)

	i := usecases.NewMockContestInteractor(ctrl)
	i.EXPECT().CreateContest(*contest).Return(nil)

	s := services.NewContestService(i)
	err := s.Create(ctx)

	assert.NoError(t, err)
}

func TestContestService_Update(t *testing.T) {
	contest := &domain.Contest{
		ID:    1,
		Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
		Open:  true,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(204)
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *contest)
	ctx.EXPECT().BindID(gomock.Any()).Return(nil)

	i := usecases.NewMockContestInteractor(ctrl)
	i.EXPECT().UpdateContest(*contest).Return(nil)

	s := services.NewContestService(i)
	err := s.Update(ctx)

	assert.NoError(t, err)
}
