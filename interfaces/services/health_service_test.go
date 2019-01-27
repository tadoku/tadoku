package services_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/interfaces/services"
)

func TestHealthService_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().String(200, "pong")

	s := services.NewHealthService()
	err := s.Ping(ctx)

	assert.NoError(t, err)
}
