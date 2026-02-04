package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockContestFindRepo struct {
	findFn func(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error)
}

func (m *mockContestFindRepo) FindContestByID(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error) {
	if m.findFn != nil {
		return m.findFn(ctx, req)
	}
	return nil, nil
}

func TestContestFind_Execute(t *testing.T) {
	t.Run("finds contest by ID successfully", func(t *testing.T) {
		contestID := uuid.New()
		ownerID := uuid.New()
		now := time.Now()

		repo := &mockContestFindRepo{
			findFn: func(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error) {
				if req.ID == contestID {
					return &domain.ContestView{
						ID:                   contestID,
						ContestStart:         now,
						ContestEnd:           now.Add(30 * 24 * time.Hour),
						RegistrationEnd:      now.Add(7 * 24 * time.Hour),
						Title:                "Test Contest",
						OwnerUserID:          ownerID,
						OwnerUserDisplayName: "Owner",
						Official:             false,
						Private:              false,
						AllowedLanguages:     []domain.Language{{Code: "ja", Name: "Japanese"}},
						AllowedActivities:    []domain.Activity{{ID: 1, Name: "Reading", Default: true}},
						CreatedAt:            now,
						UpdatedAt:            now,
					}, nil
				}
				return nil, domain.ErrNotFound
			},
		}

		svc := domain.NewContestFind(repo)
		contest, err := svc.Execute(context.Background(), &domain.ContestFindRequest{
			ID: contestID,
		})

		require.NoError(t, err)
		assert.Equal(t, contestID, contest.ID)
		assert.Equal(t, "Test Contest", contest.Title)
	})

	t.Run("sets IncludeDeleted to true for admin", func(t *testing.T) {
		contestID := uuid.New()
		var capturedReq *domain.ContestFindRequest

		repo := &mockContestFindRepo{
			findFn: func(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error) {
				capturedReq = req
				return &domain.ContestView{ID: contestID}, nil
			},
		}

		svc := domain.NewContestFind(repo)
		ctx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{Role: commondomain.RoleAdmin})
		_, err := svc.Execute(ctx, &domain.ContestFindRequest{
			ID: contestID,
		})

		require.NoError(t, err)
		assert.True(t, capturedReq.IncludeDeleted)
	})

	t.Run("sets IncludeDeleted to false for regular user", func(t *testing.T) {
		contestID := uuid.New()
		var capturedReq *domain.ContestFindRequest

		repo := &mockContestFindRepo{
			findFn: func(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error) {
				capturedReq = req
				return &domain.ContestView{ID: contestID}, nil
			},
		}

		svc := domain.NewContestFind(repo)
		ctx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{Role: commondomain.RoleUser})
		_, err := svc.Execute(ctx, &domain.ContestFindRequest{
			ID: contestID,
		})

		require.NoError(t, err)
		assert.False(t, capturedReq.IncludeDeleted)
	})

	t.Run("returns not found when contest does not exist", func(t *testing.T) {
		repo := &mockContestFindRepo{
			findFn: func(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error) {
				return nil, domain.ErrNotFound
			},
		}

		svc := domain.NewContestFind(repo)
		_, err := svc.Execute(context.Background(), &domain.ContestFindRequest{
			ID: uuid.New(),
		})

		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockContestFindRepo{
			findFn: func(ctx context.Context, req *domain.ContestFindRequest) (*domain.ContestView, error) {
				return nil, repoErr
			},
		}

		svc := domain.NewContestFind(repo)
		_, err := svc.Execute(context.Background(), &domain.ContestFindRequest{
			ID: uuid.New(),
		})

		assert.ErrorIs(t, err, repoErr)
	})
}
