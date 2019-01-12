package services

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHealthService_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := NewMockContext(ctrl)
	ctx.EXPECT().String(200, "pong")

	s := NewHealthService()
	err := s.Ping(ctx)

	assert.NoError(t, err)
}
