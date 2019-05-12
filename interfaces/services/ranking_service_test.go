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

	contestID := uint64(1)
	userID := uint64(1)

	payload := &services.CreateRankingPayload{
		ContestID: contestID,
		Languages: domain.LanguageCodes{domain.Japanese},
	}

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(201)
	ctx.EXPECT().User().Return(&domain.User{ID: userID}, nil)
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *payload)

	i := usecases.NewMockRankingInteractor(ctrl)
	i.EXPECT().CreateRanking(contestID, userID, payload.Languages).Return(nil)

	s := services.NewRankingService(i)
	err := s.Create(ctx)

	assert.NoError(t, err)
}

func TestRankingService_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	contestID := uint64(1)
	language := domain.Global

	expected := domain.Rankings{
		{ID: 1, ContestID: contestID, UserID: 1, Language: domain.Global, Amount: 15},
		{ID: 2, ContestID: contestID, UserID: 2, Language: domain.Global, Amount: 12},
		{ID: 3, ContestID: contestID, UserID: 3, Language: domain.Global, Amount: 11},
	}

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().QueryParam("contest_id").Return("1")
	ctx.EXPECT().QueryParam("language").Return(string(domain.Global))
	ctx.EXPECT().JSON(200, expected.GetView())

	i := usecases.NewMockRankingInteractor(ctrl)
	i.EXPECT().RankingsForContest(contestID, language).Return(expected, nil)

	s := services.NewRankingService(i)
	err := s.Get(ctx)

	assert.NoError(t, err)
}
