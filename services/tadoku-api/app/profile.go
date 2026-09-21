package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
)

func (a *Application) ListUsers(ctx context.Context, pageSize, page int, query string) (*profile.UserList, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.profile.ListUsers(ctx, pageSize, page, query)
}

func (a *Application) FindProfile(ctx context.Context, userID uuid.UUID) (*profile.PublicProfile, error) {
	return a.profile.FindProfile(ctx, userID)
}

func (a *Application) YearlyActivity(ctx context.Context, userID uuid.UUID, year int) (*profile.YearlyActivity, error) {
	return a.profile.YearlyActivity(ctx, userID, year)
}

func (a *Application) YearlyScores(ctx context.Context, userID uuid.UUID, year int) (*profile.YearlyScores, error) {
	return a.profile.YearlyScores(ctx, userID, year)
}

func (a *Application) YearlyActivitySplit(ctx context.Context, userID uuid.UUID, year int) ([]profile.ActivitySplitScore, error) {
	return a.profile.YearlyActivitySplit(ctx, userID, year)
}
