package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockContestCreateRepository struct {
	createCalled     bool
	createResult     *domain.ContestCreateResponse
	createErr        error
	contestCount     int32
	contestCountErr  error
	userUpsertCalled bool
}

func (m *mockContestCreateRepository) CreateContest(ctx context.Context, req *domain.ContestCreateRequest) (*domain.ContestCreateResponse, error) {
	m.createCalled = true
	return m.createResult, m.createErr
}

func (m *mockContestCreateRepository) GetContestsByUserCountForYear(ctx context.Context, now time.Time, userID uuid.UUID) (int32, error) {
	return m.contestCount, m.contestCountErr
}

func (m *mockContestCreateRepository) UpsertUser(ctx context.Context, req *domain.UserUpsertRequest) error {
	m.userUpsertCalled = true
	return nil
}

func TestContestCreate_Execute(t *testing.T) {
	now := time.Now()
	clock := commondomain.NewMockClock(now)

	validRequest := &domain.ContestCreateRequest{
		ContestStart:            now.Add(30 * 24 * time.Hour),
		ContestEnd:              now.Add(45 * 24 * time.Hour),
		RegistrationEnd:         now.Add(40 * 24 * time.Hour),
		Official:                false,
		Private:                 false,
		Title:                   "test round",
		ActivityTypeIDAllowList: []int32{1, 2},
	}

	t.Run("returns unauthorized for guest", func(t *testing.T) {
		repo := &mockContestCreateRepository{}
		userUpsert := domain.NewUserUpsert(repo)
		svc := domain.NewContestCreate(repo, clock, userUpsert)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:    commondomain.RoleGuest,
			Subject: uuid.NewString(),
		})

		_, err := svc.Execute(ctx, validRequest)

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns forbidden when non-admin creates official contest", func(t *testing.T) {
		repo := &mockContestCreateRepository{}
		userUpsert := domain.NewUserUpsert(repo)
		svc := domain.NewContestCreate(repo, clock, userUpsert)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:        commondomain.RoleUser,
			Subject:     uuid.NewString(),
			DisplayName: "TestUser",
		})

		officialRequest := &domain.ContestCreateRequest{
			ContestStart:            now.Add(30 * 24 * time.Hour),
			ContestEnd:              now.Add(45 * 24 * time.Hour),
			RegistrationEnd:         now.Add(40 * 24 * time.Hour),
			Official:                true,
			Private:                 false,
			Title:                   "test round",
			ActivityTypeIDAllowList: []int32{1, 2},
		}

		_, err := svc.Execute(ctx, officialRequest)

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, repo.createCalled)
	})

	t.Run("returns error when official and private", func(t *testing.T) {
		repo := &mockContestCreateRepository{}
		userUpsert := domain.NewUserUpsert(repo)
		svc := domain.NewContestCreate(repo, clock, userUpsert)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:        commondomain.RoleAdmin,
			Subject:     uuid.NewString(),
			DisplayName: "Admin",
		})

		badRequest := &domain.ContestCreateRequest{
			ContestStart:            now.Add(30 * 24 * time.Hour),
			ContestEnd:              now.Add(45 * 24 * time.Hour),
			RegistrationEnd:         now.Add(40 * 24 * time.Hour),
			Official:                true,
			Private:                 true,
			Title:                   "test round",
			ActivityTypeIDAllowList: []int32{1, 2},
		}

		_, err := svc.Execute(ctx, badRequest)

		assert.ErrorIs(t, err, domain.ErrInvalidContest)
		assert.False(t, repo.createCalled)
	})

	t.Run("allows user to create non-official contest", func(t *testing.T) {
		expectedResult := &domain.ContestCreateResponse{
			ID:    uuid.New(),
			Title: "test round",
		}
		repo := &mockContestCreateRepository{
			createResult: expectedResult,
			contestCount: 0,
		}
		userUpsert := domain.NewUserUpsert(repo)
		svc := domain.NewContestCreate(repo, clock, userUpsert)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:        commondomain.RoleUser,
			Subject:     uuid.NewString(),
			DisplayName: "TestUser",
		})

		result, err := svc.Execute(ctx, validRequest)

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		assert.True(t, repo.createCalled)
	})

	t.Run("returns forbidden when user exceeds yearly limit", func(t *testing.T) {
		repo := &mockContestCreateRepository{
			contestCount: 12, // At the limit
		}
		userUpsert := domain.NewUserUpsert(repo)
		svc := domain.NewContestCreate(repo, clock, userUpsert)

		ctx := ctxWithToken(&commondomain.UserIdentity{
			Role:        commondomain.RoleUser,
			Subject:     uuid.NewString(),
			DisplayName: "TestUser",
		})

		_, err := svc.Execute(ctx, validRequest)

		assert.ErrorIs(t, err, domain.ErrForbidden)
		assert.False(t, repo.createCalled)
	})
}
