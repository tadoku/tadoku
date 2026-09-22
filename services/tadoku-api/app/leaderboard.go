package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
)

type LeaderboardRequest = leaderboard.Request
type ContestLeaderboardRequest = leaderboard.ContestRequest
type YearlyLeaderboardRequest = leaderboard.YearlyRequest
type Leaderboard = leaderboard.Leaderboard

func (a *Application) FetchContestLeaderboard(ctx context.Context, request ContestLeaderboardRequest) (*Leaderboard, error) {
	result, err := a.leaderboard.FetchContest(ctx, request)
	if err != nil {
		return nil, err
	}
	return a.hydrateLeaderboard(ctx, result)
}

func (a *Application) FetchYearlyLeaderboard(ctx context.Context, request YearlyLeaderboardRequest) (*Leaderboard, error) {
	result, err := a.leaderboard.FetchYearly(ctx, request)
	if err != nil {
		return nil, err
	}
	return a.hydrateLeaderboard(ctx, result)
}

func (a *Application) FetchGlobalLeaderboard(ctx context.Context, request LeaderboardRequest) (*Leaderboard, error) {
	result, err := a.leaderboard.FetchGlobal(ctx, request)
	if err != nil {
		return nil, err
	}
	return a.hydrateLeaderboard(ctx, result)
}

func (a *Application) hydrateLeaderboard(ctx context.Context, result *leaderboard.Result) (*Leaderboard, error) {
	if !result.HydrateDisplayNames {
		return result.Leaderboard, nil
	}

	ids := make([]uuid.UUID, len(result.Leaderboard.Entries))
	for i, entry := range result.Leaderboard.Entries {
		ids[i] = entry.UserID
	}
	names, err := a.profile.DisplayNames(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch display names: %w", err)
	}
	for i := range result.Leaderboard.Entries {
		result.Leaderboard.Entries[i].UserDisplayName = names[result.Leaderboard.Entries[i].UserID]
	}
	return result.Leaderboard, nil
}
