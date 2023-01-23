package query

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

var ErrRequestInvalid = errors.New("request is invalid")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

type Repository interface {
	// contest
	FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error)
	FindByID(context.Context, *FindByIDRequest) (*ContestView, error)
	ContestFindLatestOfficial(context.Context) (*ContestView, error)
	ListContests(context.Context, *ContestListRequest) (*ContestListResponse, error)
	FindRegistrationForUser(context.Context, *FindRegistrationForUserRequest) (*ContestRegistration, error)
	FetchContestLeaderboard(context.Context, *FetchContestLeaderboardRequest) (*Leaderboard, error)
	FetchOngoingContestRegistrations(context.Context, *FetchOngoingContestRegistrationsRequest) (*ContestRegistrations, error)
	YearlyContestRegistrations(context.Context, *YearlyContestRegistrationsRequest) (*ContestRegistrations, error)

	// log
	ListLogsForContestUser(context.Context, *LogListForContestUserRequest) (*LogListResponse, error)
	FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error)
	FindLogByID(context.Context, *FindLogByIDRequest) (*Log, error)

	// profile
	FindScoresForRegistration(context.Context, *ContestProfileRequest) ([]Score, error)
	ReadingActivityForContestUser(context.Context, *ContestProfileRequest) ([]ReadingActivityRow, error)
	YearlyActivityForUser(context.Context, *YearlyActivityForUserRequest) ([]UserActivityScore, error)
	YearlyScoresForUser(context.Context, *YearlyScoresForUserRequest) ([]Score, error)
	YearlyActivitySplitForUser(context.Context, *YearlyActivitySplitForUserRequest) (*YearlyActivitySplitForUserResponse, error)
}

type Service interface {
	// contest
	FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error)
	FindByID(context.Context, *FindByIDRequest) (*ContestView, error)
	FindLatestOfficial(context.Context) (*ContestView, error)
	ListContests(context.Context, *ContestListRequest) (*ContestListResponse, error)
	FindRegistrationForUser(context.Context, *FindRegistrationForUserRequest) (*ContestRegistration, error)
	FetchContestLeaderboard(context.Context, *FetchContestLeaderboardRequest) (*Leaderboard, error)
	FetchOngoingContestRegistrations(context.Context, *FetchOngoingContestRegistrationsRequest) (*ContestRegistrations, error)
	YearlyContestRegistrations(context.Context, *YearlyContestRegistrationsRequest) (*ContestRegistrations, error)

	// log
	ListLogsForContestUser(context.Context, *LogListForContestUserRequest) (*LogListResponse, error)
	FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error)
	FindLogByID(context.Context, *FindLogByIDRequest) (*Log, error)

	// profile
	ContestProfile(context.Context, *ContestProfileRequest) (*ContestProfileResponse, error)
	// TODO: Shouldn't include reading prefix
	ReadingActivityForContestUser(context.Context, *ContestProfileRequest) (*ReadingActivityResponse, error)
	FetchUserProfile(context.Context, uuid.UUID) (*UserProfile, error)
	YearlyActivityForUser(context.Context, *YearlyActivityForUserRequest) (*YearlyActivityForUserResponse, error)
	YearlyScoresForUser(context.Context, *YearlyScoresForUserRequest) (*YearlyScoresForUserResponse, error)
	YearlyActivitySplitForUser(context.Context, *YearlyActivitySplitForUserRequest) (*YearlyActivitySplitForUserResponse, error)
}

type ServiceImpl struct {
	r        Repository
	validate *validator.Validate
	kratos   KratosClient
	clock    domain.Clock
}

type KratosClient interface {
	FetchIdentity(ctx context.Context, id uuid.UUID) (*UserTraits, error)
}

func NewService(r Repository, clock domain.Clock, kratos KratosClient) Service {
	return &ServiceImpl{
		r:        r,
		validate: validator.New(),
		clock:    clock,
		kratos:   kratos,
	}
}
