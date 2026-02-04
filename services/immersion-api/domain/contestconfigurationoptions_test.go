package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type mockContestConfigurationOptionsRepo struct {
	fetchFn func(ctx context.Context) (*domain.ContestConfigurationOptionsResponse, error)
}

func (m *mockContestConfigurationOptionsRepo) FetchContestConfigurationOptions(ctx context.Context) (*domain.ContestConfigurationOptionsResponse, error) {
	if m.fetchFn != nil {
		return m.fetchFn(ctx)
	}
	return nil, nil
}

func TestContestConfigurationOptions_Execute(t *testing.T) {
	t.Run("fetches configuration options successfully", func(t *testing.T) {
		repo := &mockContestConfigurationOptionsRepo{
			fetchFn: func(ctx context.Context) (*domain.ContestConfigurationOptionsResponse, error) {
				return &domain.ContestConfigurationOptionsResponse{
					Languages: []domain.Language{
						{Code: "ja", Name: "Japanese"},
						{Code: "zh", Name: "Chinese"},
					},
					Activities: []domain.Activity{
						{ID: 1, Name: "Reading", Default: true},
						{ID: 2, Name: "Listening", Default: false},
					},
				}, nil
			},
		}

		svc := domain.NewContestConfigurationOptions(repo)
		resp, err := svc.Execute(context.Background())

		require.NoError(t, err)
		assert.Len(t, resp.Languages, 2)
		assert.Len(t, resp.Activities, 2)
		assert.Equal(t, "ja", resp.Languages[0].Code)
		assert.Equal(t, "Reading", resp.Activities[0].Name)
		assert.False(t, resp.CanCreateOfficialRound)
	})

	t.Run("sets CanCreateOfficialRound for admin", func(t *testing.T) {
		repo := &mockContestConfigurationOptionsRepo{
			fetchFn: func(ctx context.Context) (*domain.ContestConfigurationOptionsResponse, error) {
				return &domain.ContestConfigurationOptionsResponse{
					Languages:  []domain.Language{},
					Activities: []domain.Activity{},
				}, nil
			},
		}

		svc := domain.NewContestConfigurationOptions(repo)
		ctx := context.WithValue(context.Background(), commondomain.CtxIdentityKey, &commondomain.UserIdentity{Role: commondomain.RoleAdmin})
		resp, err := svc.Execute(ctx)

		require.NoError(t, err)
		assert.True(t, resp.CanCreateOfficialRound)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockContestConfigurationOptionsRepo{
			fetchFn: func(ctx context.Context) (*domain.ContestConfigurationOptionsResponse, error) {
				return nil, repoErr
			},
		}

		svc := domain.NewContestConfigurationOptions(repo)
		_, err := svc.Execute(context.Background())

		assert.ErrorIs(t, err, repoErr)
	})
}
