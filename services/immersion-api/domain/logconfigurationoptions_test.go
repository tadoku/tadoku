package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockLogConfigurationOptionsRepo struct {
	fetchFn func(ctx context.Context, userID uuid.UUID) (*domain.LogConfigurationOptionsResponse, error)
}

func (m *mockLogConfigurationOptionsRepo) FetchLogConfigurationOptions(ctx context.Context, userID uuid.UUID) (*domain.LogConfigurationOptionsResponse, error) {
	if m.fetchFn != nil {
		return m.fetchFn(ctx, userID)
	}
	return nil, nil
}

func TestLogConfigurationOptions_Execute(t *testing.T) {
	t.Run("fetches configuration options successfully for authenticated user", func(t *testing.T) {
		unitID := uuid.New()
		langCode := "ja"

		repo := &mockLogConfigurationOptionsRepo{
			fetchFn: func(ctx context.Context, userID uuid.UUID) (*domain.LogConfigurationOptionsResponse, error) {
				return &domain.LogConfigurationOptionsResponse{
					Languages: []domain.Language{
						{Code: "ja", Name: "Japanese"},
					},
					Activities: []domain.Activity{
						{ID: 1, Name: "Reading", Default: true},
					},
					Units: []domain.Unit{
						{ID: unitID, LogActivityID: 1, Name: "Pages", Modifier: 1.0, LanguageCode: &langCode},
					},
					UserLanguageCodes: []string{"ja"},
				}, nil
			},
		}

		svc := domain.NewLogConfigurationOptions(repo)
		ctx := ctxWithUser()
		resp, err := svc.Execute(ctx)

		require.NoError(t, err)
		assert.Len(t, resp.Languages, 1)
		assert.Len(t, resp.Activities, 1)
		assert.Len(t, resp.Units, 1)
		assert.Equal(t, "ja", resp.Languages[0].Code)
		assert.Equal(t, "Reading", resp.Activities[0].Name)
		assert.Equal(t, "Pages", resp.Units[0].Name)
		assert.Equal(t, []string{"ja"}, resp.UserLanguageCodes)
	})

	t.Run("returns unauthorized for guest user", func(t *testing.T) {
		repo := &mockLogConfigurationOptionsRepo{}

		svc := domain.NewLogConfigurationOptions(repo)
		ctx := ctxWithGuest()
		_, err := svc.Execute(ctx)

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("returns unauthorized when no session", func(t *testing.T) {
		repo := &mockLogConfigurationOptionsRepo{
			fetchFn: func(ctx context.Context, userID uuid.UUID) (*domain.LogConfigurationOptionsResponse, error) {
				return &domain.LogConfigurationOptionsResponse{}, nil
			},
		}

		svc := domain.NewLogConfigurationOptions(repo)
		_, err := svc.Execute(context.Background())

		assert.ErrorIs(t, err, domain.ErrUnauthorized)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockLogConfigurationOptionsRepo{
			fetchFn: func(ctx context.Context, userID uuid.UUID) (*domain.LogConfigurationOptionsResponse, error) {
				return nil, repoErr
			},
		}

		svc := domain.NewLogConfigurationOptions(repo)
		ctx := ctxWithUser()
		_, err := svc.Execute(ctx)

		assert.ErrorIs(t, err, repoErr)
	})
}
