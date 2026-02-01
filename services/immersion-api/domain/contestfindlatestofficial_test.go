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

type mockContestFindLatestOfficialRepo struct {
	findFn func(ctx context.Context) (*domain.ContestView, error)
}

func (m *mockContestFindLatestOfficialRepo) ContestFindLatestOfficial(ctx context.Context) (*domain.ContestView, error) {
	if m.findFn != nil {
		return m.findFn(ctx)
	}
	return nil, nil
}

func TestContestFindLatestOfficial_Execute(t *testing.T) {
	t.Run("finds latest official contest successfully", func(t *testing.T) {
		contestID := uuid.New()
		ownerID := uuid.New()
		now := time.Now()

		repo := &mockContestFindLatestOfficialRepo{
			findFn: func(ctx context.Context) (*domain.ContestView, error) {
				return &domain.ContestView{
					ID:                   contestID,
					ContestStart:         now,
					ContestEnd:           now.Add(30 * 24 * time.Hour),
					RegistrationEnd:      now.Add(7 * 24 * time.Hour),
					Title:                "Official Round 2024",
					OwnerUserID:          ownerID,
					OwnerUserDisplayName: "Admin",
					Official:             true,
					Private:              false,
					AllowedLanguages:     []domain.Language{{Code: "ja", Name: "Japanese"}},
					AllowedActivities:    []domain.Activity{{ID: 1, Name: "Reading", Default: true}},
					CreatedAt:            now,
					UpdatedAt:            now,
				}, nil
			},
		}

		svc := domain.NewContestFindLatestOfficial(repo)
		contest, err := svc.Execute(context.Background())

		require.NoError(t, err)
		assert.Equal(t, contestID, contest.ID)
		assert.Equal(t, "Official Round 2024", contest.Title)
		assert.True(t, contest.Official)
		assert.Len(t, contest.AllowedLanguages, 1)
	})

	t.Run("returns not found when no official contest exists", func(t *testing.T) {
		repo := &mockContestFindLatestOfficialRepo{
			findFn: func(ctx context.Context) (*domain.ContestView, error) {
				return nil, domain.ErrNotFound
			},
		}

		svc := domain.NewContestFindLatestOfficial(repo)
		_, err := svc.Execute(context.Background())

		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockContestFindLatestOfficialRepo{
			findFn: func(ctx context.Context) (*domain.ContestView, error) {
				return nil, repoErr
			},
		}

		svc := domain.NewContestFindLatestOfficial(repo)
		_, err := svc.Execute(context.Background())

		assert.ErrorIs(t, err, repoErr)
	})
}
