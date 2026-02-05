package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockUserUpsertRepository struct {
	lastRequest *domain.UserUpsertRequest
	err         error
}

func (m *mockUserUpsertRepository) UpsertUser(ctx context.Context, req *domain.UserUpsertRequest) error {
	m.lastRequest = req
	return m.err
}

func TestUserUpsert_Execute(t *testing.T) {
	t.Run("returns unauthorized for nil session", func(t *testing.T) {
		repo := &mockUserUpsertRepository{}
		svc := domain.NewUserUpsert(repo)

		err := svc.Execute(context.Background())

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("upserts user from session", func(t *testing.T) {
		repo := &mockUserUpsertRepository{}
		svc := domain.NewUserUpsert(repo)

		userID := uuid.New()
		createdAt := time.Now()
		ctx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{
			Subject:     userID.String(),
			DisplayName: "TestUser",
			CreatedAt:   createdAt,
		})

		err := svc.Execute(ctx)

		require.NoError(t, err)
		require.NotNil(t, repo.lastRequest)
		assert.Equal(t, userID, repo.lastRequest.ID)
		assert.Equal(t, "TestUser", repo.lastRequest.DisplayName)
		assert.Equal(t, createdAt, repo.lastRequest.SessionCreatedAt)
	})
}
