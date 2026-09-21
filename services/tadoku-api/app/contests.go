package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type ListContestsParameters = contests.ListParameters
type ContestView = contests.ContestView
type Contest = contests.Contest
type CreateContestParameters = contests.CreateContestParameters

var ErrAccountDeletionInProgress = profile.ErrAccountDeletionInProgress

func (a *Application) CheckContestCreatePermission(ctx context.Context) error {
	if a.permissions.IsAdminOrFalse(ctx) {
		return nil
	}

	userID, err := identity.FromContext(ctx).UUID()
	if err != nil {
		return err
	}
	return a.contests.CheckCreatePermission(ctx, userID)
}

func (a *Application) CreateContest(ctx context.Context, parameters CreateContestParameters) (*Contest, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	admin, err := a.permissions.IsAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if parameters.Official && !admin {
		return nil, contests.ErrContestCreationForbidden
	}

	creator := identity.FromContext(ctx)
	now := timex.Now()
	creatorID, err := a.profile.SynchronizeUser(ctx, creator, now)
	if err != nil {
		return nil, err
	}

	if err := a.contests.ValidateContestCreation(ctx, parameters, creatorID, creator.DisplayName, admin, now); err != nil {
		return nil, err
	}
	prepared := a.contests.PrepareContestCreation(parameters, creatorID, creator.DisplayName, now)

	var result *contests.Contest
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, creatorID); err != nil {
			return err
		}
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
