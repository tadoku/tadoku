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
type ContestRegistration = contests.Registration
type ContestRegistrationList = contests.RegistrationList
type ContestRegistrationUpsertParameters = contests.RegistrationUpsertParameters

var ErrAccountDeletionInProgress = profile.ErrAccountDeletionInProgress
var ErrContestRegistrationNotFound = contests.ErrRegistrationNotFound

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
	contest := contests.Contest{
		ID:                      uuid.New(),
		ContestStart:            parameters.ContestStart,
		ContestEnd:              parameters.ContestEnd,
		RegistrationEnd:         parameters.RegistrationEnd,
		Title:                   parameters.Title,
		Description:             parameters.Description,
		OwnerUserID:             creatorID,
		OwnerUserDisplayName:    creator.DisplayName,
		Official:                parameters.Official,
		Private:                 parameters.Private,
		LanguageCodeAllowList:   parameters.LanguageCodeAllowList,
		ActivityTypeIDAllowList: parameters.ActivityTypeIDAllowList,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	var result *contests.Contest
	err = postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, creatorID); err != nil {
			return err
		}
		var createErr error
		result, createErr = a.contests.CreateContest(ctx, contest)
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

func (a *Application) FindContestRegistration(ctx context.Context, contestID uuid.UUID) (*ContestRegistration, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	user, err := a.profile.SignedUser(ctx)
	if err != nil {
		return nil, err
	}
	return a.contests.FindRegistration(ctx, user.ID, contestID)
}

func (a *Application) ListOngoingContestRegistrations(ctx context.Context) (*ContestRegistrationList, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	user, err := a.profile.SignedUser(ctx)
	if err != nil {
		return nil, err
	}
	return a.contests.ListOngoingRegistrations(ctx, user.ID)
}

func (a *Application) UpsertContestRegistration(ctx context.Context, parameters ContestRegistrationUpsertParameters) error {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}
	user, err := a.profile.SignedUser(ctx)
	if err != nil {
		return err
	}
	now := timex.Now()
	if err := a.profile.SynchronizeUser(ctx, user, now); err != nil {
		return err
	}
	prepared, err := a.contests.PrepareRegistrationUpsert(ctx, parameters, user.ID, now)
	if err != nil {
		return err
	}
	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, user.ID); err != nil {
			return err
		}
		return a.contests.ApplyRegistration(ctx, prepared)
	})
}
