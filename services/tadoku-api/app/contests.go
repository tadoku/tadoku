package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type ListContestsParameters = contests.ListParameters
type ContestView = contests.ContestView

var ErrInvalidContestCreator = contests.ErrInvalidContestCreator

func (a *Application) CheckContestCreatePermission(ctx context.Context) error {
	if a.permissions.IsAdminOrFalse(ctx) {
		return nil
	}

	user := identity.FromContext(ctx)
	if user == nil {
		return contests.ErrInvalidContestCreator
	}
	userID, err := uuid.Parse(user.Subject)
	if err != nil {
		return contests.ErrInvalidContestCreator
	}
	return a.contests.CheckCreatePermission(ctx, userID)
}

func (a *Application) ListContests(ctx context.Context, parameters ListContestsParameters) (*contests.ContestList, error) {
	includePrivate := a.permissions.IsAdminOrFalse(ctx)

	return a.contests.ListContests(ctx, parameters, includePrivate)
}

func (a *Application) FindContestByID(ctx context.Context, id uuid.UUID) (*contests.ContestView, error) {
	includeDeleted := a.permissions.IsAdminOrFalse(ctx)
	return a.contests.FindContestByID(ctx, id, includeDeleted)
}

func (a *Application) FindLatestOfficialContest(ctx context.Context) (*contests.ContestView, error) {
	return a.contests.FindLatestOfficialContest(ctx)
}

func (a *Application) ContestConfigurationOptions(ctx context.Context) (*contests.ConfigurationOptions, error) {
	canCreateOfficialRound := a.permissions.IsAdminOrFalse(ctx)
	return a.contests.ConfigurationOptions(ctx, canCreateOfficialRound)
}
