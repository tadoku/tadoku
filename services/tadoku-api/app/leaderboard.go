package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
)

type LeaderboardRequest = leaderboard.Request
type ContestLeaderboardRequest = leaderboard.ContestRequest
type YearlyLeaderboardRequest = leaderboard.YearlyRequest
type Leaderboard = leaderboard.Leaderboard

func (a *Application) FetchContestLeaderboard(ctx context.Context, request ContestLeaderboardRequest) (*Leaderboard, error) {
	return a.leaderboard.FetchContest(ctx, request)
}

func (a *Application) FetchYearlyLeaderboard(ctx context.Context, request YearlyLeaderboardRequest) (*Leaderboard, error) {
	return a.leaderboard.FetchYearly(ctx, request)
}

func (a *Application) FetchGlobalLeaderboard(ctx context.Context, request LeaderboardRequest) (*Leaderboard, error) {
	return a.leaderboard.FetchGlobal(ctx, request)
}
