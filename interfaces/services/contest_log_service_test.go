package services_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/interfaces/services"
	"github.com/tadoku/api/usecases"
)

func TestContestLogService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := &domain.ContestLog{
		ContestID: 1,
		Language:  domain.Japanese,
		Amount:    10,
		MediumID:  1,
	}

	ctx := services.NewMockContext(ctrl)
	ctx.EXPECT().NoContent(201)
	ctx.EXPECT().User().Return(&domain.User{ID: 1}, nil)
	ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *log)

	i := usecases.NewMockRankingInteractor(ctrl)
	i.EXPECT().CreateLog(domain.ContestLog{ContestID: 1, UserID: 1, Language: domain.Japanese, Amount: 10, MediumID: 1}).Return(nil)

	s := services.NewContestLogService(i)
	err := s.Create(ctx)

	assert.NoError(t, err)
}

func TestContestLogService_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	{
		log := &domain.ContestLog{
			ContestID: 1,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  1,
		}

		ctx := services.NewMockContext(ctrl)
		ctx.EXPECT().NoContent(204)
		ctx.EXPECT().User().Return(&domain.User{ID: 1}, nil)
		ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *log)
		ctx.EXPECT().BindID(gomock.Any()).Return(nil).SetArg(0, uint64(1))

		i := usecases.NewMockRankingInteractor(ctrl)
		i.EXPECT().UpdateLog(domain.ContestLog{ID: 1, ContestID: 1, UserID: 1, Language: domain.Japanese, Amount: 10, MediumID: 1}).Return(nil)

		s := services.NewContestLogService(i)
		err := s.Update(ctx)

		assert.NoError(t, err)
	}

	{
		log := &domain.ContestLog{
			ContestID: 1,
			Language:  domain.Japanese,
			Amount:    10,
			MediumID:  1,
		}

		ctx := services.NewMockContext(ctrl)
		ctx.EXPECT().NoContent(403)
		ctx.EXPECT().User().Return(&domain.User{ID: 1}, nil)
		ctx.EXPECT().Bind(gomock.Any()).Return(nil).SetArg(0, *log)
		ctx.EXPECT().BindID(gomock.Any()).Return(nil).SetArg(0, uint64(1))

		i := usecases.NewMockRankingInteractor(ctrl)
		i.EXPECT().UpdateLog(domain.ContestLog{ID: 1, ContestID: 1, UserID: 1, Language: domain.Japanese, Amount: 10, MediumID: 1}).Return(domain.ErrInsufficientPermissions)

		s := services.NewContestLogService(i)
		err := s.Update(ctx)

		assert.NoError(t, err)
	}
}

func TestContestLogService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logID := uint64(1)
	userID := uint64(2)

	// Happy path
	{
		ctx := services.NewMockContext(ctrl)
		ctx.EXPECT().NoContent(200)
		ctx.EXPECT().User().Return(&domain.User{ID: userID}, nil)
		ctx.EXPECT().BindID(gomock.Any()).Return(nil).SetArg(0, logID)

		i := usecases.NewMockRankingInteractor(ctrl)
		i.EXPECT().DeleteLog(logID, userID).Return(nil)

		s := services.NewContestLogService(i)
		err := s.Delete(ctx)

		assert.NoError(t, err)
	}

	// Sad path: log is not the user's
	{
		ctx := services.NewMockContext(ctrl)
		ctx.EXPECT().NoContent(403)
		ctx.EXPECT().User().Return(&domain.User{ID: userID}, nil)
		ctx.EXPECT().BindID(gomock.Any()).Return(nil).SetArg(0, logID)

		i := usecases.NewMockRankingInteractor(ctrl)
		i.EXPECT().DeleteLog(logID, userID).Return(domain.ErrInsufficientPermissions)

		s := services.NewContestLogService(i)
		err := s.Delete(ctx)

		assert.NoError(t, err)
	}
}
