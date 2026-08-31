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

type recordingLeaderboardCacheObserver struct {
	observations []domain.LeaderboardCacheObservation
}

func (o *recordingLeaderboardCacheObserver) ObserveLeaderboardCache(
	_ context.Context,
	observation domain.LeaderboardCacheObservation,
) {
	o.observations = append(o.observations, observation)
}

func TestLeaderboardReadsObservePostgresFallbacks(t *testing.T) {
	cacheErr := errors.New("valkey unavailable")

	tests := []struct {
		name string
		kind domain.LeaderboardCacheKind
		run  func(*recordingLeaderboardCacheObserver) error
	}{
		{
			name: "global",
			kind: domain.LeaderboardCacheKindGlobal,
			run: func(observer *recordingLeaderboardCacheObserver) error {
				service := domain.NewLeaderboardGlobalWithCacheObserver(
					&leaderboardGlobalRepositoryMock{leaderboard: &domain.Leaderboard{}},
					&leaderboardGlobalStoreMock{fetchErr: cacheErr},
					observer,
				)
				_, err := service.Execute(context.Background(), &domain.LeaderboardGlobalRequest{PageSize: 25})
				return err
			},
		},
		{
			name: "yearly",
			kind: domain.LeaderboardCacheKindYearly,
			run: func(observer *recordingLeaderboardCacheObserver) error {
				service := domain.NewLeaderboardYearlyWithCacheObserver(
					&leaderboardYearlyRepositoryMock{leaderboard: &domain.Leaderboard{}},
					&leaderboardYearlyStoreMock{fetchErr: cacheErr},
					observer,
				)
				_, err := service.Execute(context.Background(), &domain.LeaderboardYearlyRequest{Year: 2026, PageSize: 25})
				return err
			},
		},
		{
			name: "contest",
			kind: domain.LeaderboardCacheKindContest,
			run: func(observer *recordingLeaderboardCacheObserver) error {
				service := domain.NewContestLeaderboardFetchWithCacheObserver(
					&contestLeaderboardFetchRepositoryMock{leaderboard: &domain.Leaderboard{}},
					&contestLeaderboardFetchStoreMock{fetchErr: cacheErr},
					observer,
				)
				_, err := service.Execute(context.Background(), &domain.ContestLeaderboardFetchRequest{
					ContestID: uuid.New(),
					PageSize:  25,
				})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observer := &recordingLeaderboardCacheObserver{}

			require.NoError(t, test.run(observer))
			require.Len(t, observer.observations, 1)
			assert.Equal(t, test.kind, observer.observations[0].Kind)
			assert.Equal(t, domain.LeaderboardCacheOperationFetch, observer.observations[0].Operation)
			assert.Equal(t, domain.LeaderboardCacheOutcomeFallback, observer.observations[0].Outcome)
			assert.ErrorIs(t, observer.observations[0].Err, cacheErr)
		})
	}
}
