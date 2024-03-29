package query

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

var ErrRequestInvalid = errors.New("request is invalid")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")
var ErrForbidden = errors.New("forbidden")

type Repository interface {
	// contest
	FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error)
	FindContestByID(context.Context, *FindContestByIDRequest) (*ContestView, error)
	ContestFindLatestOfficial(context.Context) (*ContestView, error)
	ListContests(context.Context, *ListContestsRequest) (*ListContestsResponse, error)
	FindRegistrationForUser(context.Context, *FindRegistrationForUserRequest) (*ContestRegistration, error)
	FetchContestLeaderboard(context.Context, *FetchContestLeaderboardRequest) (*Leaderboard, error)
	FetchOngoingContestRegistrations(context.Context, *FetchOngoingContestRegistrationsRequest) (*ContestRegistrations, error)
	YearlyContestRegistrationsForUser(context.Context, *YearlyContestRegistrationsForUserRequest) (*ContestRegistrations, error)
	FetchYearlyLeaderboard(context.Context, *FetchYearlyLeaderboardRequest) (*Leaderboard, error)
	FetchGlobalLeaderboard(context.Context, *FetchGlobalLeaderboardRequest) (*Leaderboard, error)
	FetchContestSummary(context.Context, *FetchContestSummaryRequest) (*FetchContestSummaryResponse, error)
	GetContestsByUserCountForYear(context.Context, time.Time, uuid.UUID) (int32, error)

	// log
	ListLogsForUser(context.Context, *ListLogsForUserRequest) (*ListLogsForUserResponse, error)
	ListLogsForContest(context.Context, *ListLogsForContestRequest) (*ListLogsForContestResponse, error)
	FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error)
	FindLogByID(context.Context, *FindLogByIDRequest) (*Log, error)

	// profile
	FindScoresForRegistration(context.Context, *ContestProfileRequest) ([]Score, error)
	ActivityForContestUser(context.Context, *ActivityForContestUserRequest) ([]ActivityForContestUserRow, error)
	YearlyActivityForUser(context.Context, *YearlyActivityForUserRequest) ([]UserActivityScore, error)
	YearlyScoresForUser(context.Context, *YearlyScoresForUserRequest) ([]Score, error)
	YearlyActivitySplitForUser(context.Context, *YearlyActivitySplitForUserRequest) (*YearlyActivitySplitForUserResponse, error)
}

type Service interface {
	// contest
	FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error)
	FindContestByID(context.Context, *FindContestByIDRequest) (*ContestView, error)
	FindLatestOfficial(context.Context) (*ContestView, error)
	ListContests(context.Context, *ListContestsRequest) (*ListContestsResponse, error)
	FindRegistrationForUser(context.Context, *FindRegistrationForUserRequest) (*ContestRegistration, error)
	FetchContestLeaderboard(context.Context, *FetchContestLeaderboardRequest) (*Leaderboard, error)
	FetchOngoingContestRegistrations(context.Context, *FetchOngoingContestRegistrationsRequest) (*ContestRegistrations, error)
	YearlyContestRegistrationsForUser(context.Context, *YearlyContestRegistrationsForUserRequest) (*ContestRegistrations, error)
	FetchYearlyLeaderboard(context.Context, *FetchYearlyLeaderboardRequest) (*Leaderboard, error)
	FetchGlobalLeaderboard(context.Context, *FetchGlobalLeaderboardRequest) (*Leaderboard, error)
	FetchContestSummary(context.Context, *FetchContestSummaryRequest) (*FetchContestSummaryResponse, error)
	CreateContestPermissionCheck(context.Context) error

	// log
	ListLogsForUser(context.Context, *ListLogsForUserRequest) (*ListLogsForUserResponse, error)
	ListLogsForContest(context.Context, *ListLogsForContestRequest) (*ListLogsForContestResponse, error)
	FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error)
	FindLogByID(context.Context, *FindLogByIDRequest) (*Log, error)

	// profile
	ContestProfile(context.Context, *ContestProfileRequest) (*ContestProfileResponse, error)
	ActivityForContestUser(context.Context, *ActivityForContestUserRequest) (*ActivityForContestUserResponse, error)
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
