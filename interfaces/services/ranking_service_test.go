package services_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/services"
	"github.com/tadoku/api/usecases"
)

func TestRankingService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	payload := &services.CreateRankingPayload{
		ContestID: 1,
		Languages: domain.LanguageCodes{domain.Japanese},
	}

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(201)
	ctx.EXPECT().User().Return(&domain.User{ID: 1}, nil)
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *payload)

	i := usecases.NewMockRankingInteractor(ctrl)
	i.EXPECT().CreateRanking(uint64(1), uint64(1), payload.Languages).Return(nil)

	s := services.NewRankingService(i)
	err := s.Create(ctx)

	assert.NoError(t, err)
}
