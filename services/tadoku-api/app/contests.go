package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type ListContestsParameters = contests.ListParameters
type ContestView = contests.ContestView
type Contest = contests.Contest
type CreateContestParameters = contests.CreateParameters

var ErrInvalidContestCreator = contests.ErrInvalidContestCreator
var ErrAccountDeletionInProgress = contests.ErrAccountDeletionInProgress

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

func (a *Application) CreateContest(ctx context.Context, parameters CreateContestParameters) (*Contest, error) {
	admin := false
	if parameters.Official {
		if err := a.permissions.RequireAdmin(ctx); err != nil {
			return nil, err
		}
		admin = true
	} else {
		if err := a.permissions.RequireAuthenticated(ctx); err != nil {
			return nil, err
		}
		var err error
		admin, err = a.permissions.IsAdmin(ctx)
		if err != nil {
			return nil, err
		}
	}

	prepared, err := a.contests.PrepareContestCreation(ctx, parameters, admin)
	if err != nil {
		return nil, err
	}

	var result *contests.Contest
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var createErr error
		result, createErr = a.contests.CreateContest(ctx, prepared)
		return createErr
	})
	if err != nil {
		result = nil
		return nil, err
	}
	return result, nil
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
