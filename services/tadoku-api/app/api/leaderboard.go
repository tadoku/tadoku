package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
)

type LeaderboardRequest = leaderboard.Request
type ContestLeaderboardRequest = leaderboard.ContestRequest
type YearlyLeaderboardRequest = leaderboard.YearlyRequest
type ContestLeaderboardEntryRequest = leaderboard.ContestEntryRequest
type Leaderboard = leaderboard.Leaderboard
type LeaderboardEntry = leaderboard.Entry

var ErrLeaderboardEntryNotFound = leaderboard.ErrEntryNotFound

func (a *Application) FetchContestLeaderboard(ctx context.Context, request ContestLeaderboardRequest) (*Leaderboard, error) {
	board, err := a.leaderboard.FetchContest(ctx, request)
	if err != nil {
		return nil, err
	}
	return a.hydrateLeaderboard(ctx, board)
}

func (a *Application) FetchContestLeaderboardEntry(
	ctx context.Context,
	request ContestLeaderboardEntryRequest,
) (*LeaderboardEntry, error) {
	entry, err := a.leaderboard.FetchContestEntry(ctx, request)
	if err != nil {
		return nil, err
	}

	board, err := a.hydrateLeaderboard(ctx, &Leaderboard{Entries: []LeaderboardEntry{*entry}})
	if err != nil {
		return nil, err
	}
	return &board.Entries[0], nil
}

func (a *Application) FetchYearlyLeaderboard(ctx context.Context, request YearlyLeaderboardRequest) (*Leaderboard, error) {
	board, err := a.leaderboard.FetchYearly(ctx, request)
	if err != nil {
		return nil, err
	}
	return a.hydrateLeaderboard(ctx, board)
}

func (a *Application) FetchGlobalLeaderboard(ctx context.Context, request LeaderboardRequest) (*Leaderboard, error) {
	board, err := a.leaderboard.FetchGlobal(ctx, request)
	if err != nil {
		return nil, err
	}
	return a.hydrateLeaderboard(ctx, board)
}

func (a *Application) hydrateLeaderboard(ctx context.Context, board *Leaderboard) (*Leaderboard, error) {
	ids := make([]uuid.UUID, len(board.Entries))
	for i, entry := range board.Entries {
		ids[i] = entry.UserID
	}
	names, err := a.profile.DisplayNames(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch display names: %w", err)
	}
	for i := range board.Entries {
		board.Entries[i].UserDisplayName = names[board.Entries[i].UserID]
	}
	return board, nil
}
