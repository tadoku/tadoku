package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/activities"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type ListContestsParameters = contests.ListParameters
type ContestView = contests.ContestView
type Contest = contests.Contest
type ContestSummary = contests.ContestSummary
type CreateContestParameters = contests.CreateContestParameters
type ContestRegistration = contests.Registration
type ContestRegistrationList = contests.RegistrationList
type ContestRegistrationUpsertParameters = contests.RegistrationUpsertParameters

var ErrAccountDeletionInProgress = profile.ErrAccountDeletionInProgress
var ErrContestRegistrationNotFound = contests.ErrRegistrationNotFound

func (a *Application) CheckContestCreatePermission(ctx context.Context) error {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}
	if a.permissions.IsAdminOrFalse(ctx) {
		return nil
	}

	userID, err := identity.RequireActorID(ctx)
	if err != nil {
		return err
	}

	accountCreatedAt, err := a.profile.FetchAccountCreatedAt(ctx, userID)
	if err != nil {
		return err
	}

	return a.contests.CheckCreatePermission(ctx, userID, accountCreatedAt)
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
	if err := a.languages.RequireExistingLanguages(ctx, parameters.LanguageCodeAllowList); err != nil {
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

func (a *Application) FetchContestSummary(ctx context.Context, id uuid.UUID) (*ContestSummary, error) {
	return a.contests.FetchContestSummary(ctx, id)
}

type ContestConfigurationOptions struct {
	Languages              []languages.Language
	Activities             []activities.Activity
	CanCreateOfficialRound bool
}

func (a *Application) ContestConfigurationOptions(ctx context.Context) (*ContestConfigurationOptions, error) {
	canCreateOfficialRound := a.permissions.IsAdminOrFalse(ctx)

	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return &ContestConfigurationOptions{
		Languages:              languages,
		Activities:             activities.All(),
		CanCreateOfficialRound: canCreateOfficialRound,
	}, nil
}

func (a *Application) FindContestRegistration(ctx context.Context, contestID uuid.UUID) (*ContestRegistration, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	userID, err := identity.RequireActorID(ctx)
	if err != nil {
		return nil, err
	}

	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return a.contests.FindRegistration(ctx, userID, contestID, languages)
}

func (a *Application) ListOngoingContestRegistrations(ctx context.Context) (*ContestRegistrationList, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	userID, err := identity.RequireActorID(ctx)
	if err != nil {
		return nil, err
	}

	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return a.contests.ListOngoingRegistrations(ctx, userID, languages)
}

func (a *Application) UpsertContestRegistration(ctx context.Context, parameters ContestRegistrationUpsertParameters) error {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}

	user := identity.FromContext(ctx)
	now := timex.Now()
	userID, err := a.profile.SynchronizeUser(ctx, user, now)
	if err != nil {
		return err
	}

	registrationID := uuid.New()
	contest, err := a.contests.ValidateRegistrationUpsert(ctx, parameters)
	if err != nil {
		return err
	}
	if err := a.languages.RequireExistingLanguages(ctx, parameters.LanguageCodes); err != nil {
		return err
	}

	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return err
	}

	existing, err := a.contests.FindRegistration(ctx, userID, parameters.ContestID, languages)
	if errors.Is(err, contests.ErrRegistrationNotFound) {
		existing = nil
	} else if err != nil {
		return err
	}

	if existing != nil {
		registrationID = existing.ID
	}

	registration := contests.Registration{
		ID:            registrationID,
		ContestID:     parameters.ContestID,
		UserID:        userID,
		LanguageCodes: parameters.LanguageCodes,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, userID); err != nil {
			return err
		}

		return a.contests.ApplyRegistration(ctx, registration, existing, *contest)
	})
}

func (a *Application) ListYearlyContestRegistrations(ctx context.Context, userID uuid.UUID, year int) (*ContestRegistrationList, error) {
	if identity.FromContext(ctx) == nil {
		return nil, errx.NewUnauthorizedError("unauthorized")
	}

	actorID, ok := identity.ActorID(ctx)
	includePrivate := a.permissions.IsAdminOrFalse(ctx) || (ok && actorID == userID)

	languages, err := a.languages.ListLanguages(ctx)
	if err != nil {
		return nil, err
	}

	return a.contests.ListYearlyRegistrations(ctx, userID, year, includePrivate, languages)
}
