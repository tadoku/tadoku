package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockProfileFetchKratos struct {
	traits *domain.UserTraits
	err    error
}

func (m *mockProfileFetchKratos) FetchIdentity(ctx context.Context, id uuid.UUID) (*domain.UserTraits, error) {
	return m.traits, m.err
}

func TestProfileFetch_Execute(t *testing.T) {
	userID := uuid.New()
	createdAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	t.Run("returns user profile on success", func(t *testing.T) {
		kratos := &mockProfileFetchKratos{
			traits: &domain.UserTraits{
				UserDisplayName: "TestUser",
				Email:           "test@example.com",
				CreatedAt:       createdAt,
			},
		}
		svc := domain.NewProfileFetch(kratos)

		profile, err := svc.Execute(context.Background(), &domain.ProfileFetchRequest{
			UserID: userID,
		})

		require.NoError(t, err)
		assert.Equal(t, "TestUser", profile.DisplayName)
		assert.Equal(t, createdAt, profile.CreatedAt)
	})

	t.Run("returns error when kratos fails", func(t *testing.T) {
		kratos := &mockProfileFetchKratos{
			err: errors.New("identity not found"),
		}
		svc := domain.NewProfileFetch(kratos)

		profile, err := svc.Execute(context.Background(), &domain.ProfileFetchRequest{
			UserID: userID,
		})

		assert.Nil(t, profile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not fetch user profile")
	})
}
