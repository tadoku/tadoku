package services_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/interfaces/services"
)

func TestRankingService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(201)

	s := services.NewRankingService()
	err := s.Create(ctx)

	assert.NoError(t, err)
}
