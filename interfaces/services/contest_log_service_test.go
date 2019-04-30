package services_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/interfaces/services"
	"github.com/tadoku/api/usecases"
)

func TestContestLogService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(201)

	i := usecases.NewMockContestLogInteractor(ctrl)

	s := services.NewContestLogService(i)
	err := s.Create(ctx)

	assert.NoError(t, err)
}
