package query

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
)

var ErrRequestInvalid = errors.New("request is invalid")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")
var ErrForbidden = errors.New("forbidden")

type Repository interface {
	// contest
	FetchContestConfigurationOptions(ctx context.Context) (*immersiondomain.ContestConfigurationOptionsResponse, error)
	FindContestByID(context.Context, *immersiondomain.ContestFindRequest) (*immersiondomain.ContestView, error)
	ContestFindLatestOfficial(context.Context) (*immersiondomain.ContestView, error)
	ListContests(context.Context, *immersiondomain.ContestListRequest) (*immersiondomain.ContestListResponse, error)
	FindRegistrationForUser(context.Context, *FindRegistrationForUserRequest) (*ContestRegistration, error)
	FetchContestLeaderboard(context.Context, *FetchContestLeaderboardRequest) (*Leaderboard, error)
	FetchOngoingContestRegistrations(context.Context, *FetchOngoingContestRegistrationsRequest) (*ContestRegistrations, error)
	YearlyContestRegistrationsForUser(context.Context, *YearlyContestRegistrationsForUserRequest) (*ContestRegistrations, error)
	FetchYearlyLeaderboard(context.Context, *FetchYearlyLeaderboardRequest) (*Leaderboard, error)
	FetchGlobalLeaderboard(context.Context, *FetchGlobalLeaderboardRequest) (*Leaderboard, error)
	FetchContestSummary(context.Context, *immersiondomain.ContestSummaryFetchRequest) (*immersiondomain.ContestSummaryFetchResponse, error)
	GetContestsByUserCountForYear(context.Context, time.Time, uuid.UUID) (int32, error)

	// log
	ListLogsForUser(context.Context, *ListLogsForUserRequest) (*ListLogsForUserResponse, error)
	ListLogsForContest(context.Context, *ListLogsForContestRequest) (*ListLogsForContestResponse, error)
	FetchLogConfigurationOptions(ctx context.Context) (*immersiondomain.LogConfigurationOptionsResponse, error)
	FindLogByID(context.Context, *immersiondomain.LogFindRequest) (*immersiondomain.Log, error)

	// profile
	FindScoresForRegistration(context.Context, *ContestProfileRequest) ([]Score, error)
	ActivityForContestUser(context.Context, *ActivityForContestUserRequest) ([]ActivityForContestUserRow, error)
	YearlyActivityForUser(context.Context, *YearlyActivityForUserRequest) ([]UserActivityScore, error)
	YearlyScoresForUser(context.Context, *YearlyScoresForUserRequest) ([]Score, error)
	YearlyActivitySplitForUser(context.Context, *immersiondomain.ProfileYearlyActivitySplitRequest) (*immersiondomain.ProfileYearlyActivitySplitResponse, error)
}

type Service interface {
	// contest
	FindRegistrationForUser(context.Context, *FindRegistrationForUserRequest) (*ContestRegistration, error)
	FetchContestLeaderboard(context.Context, *FetchContestLeaderboardRequest) (*Leaderboard, error)
	FetchOngoingContestRegistrations(context.Context, *FetchOngoingContestRegistrationsRequest) (*ContestRegistrations, error)
	YearlyContestRegistrationsForUser(context.Context, *YearlyContestRegistrationsForUserRequest) (*ContestRegistrations, error)
	FetchYearlyLeaderboard(context.Context, *FetchYearlyLeaderboardRequest) (*Leaderboard, error)
	FetchGlobalLeaderboard(context.Context, *FetchGlobalLeaderboardRequest) (*Leaderboard, error)
	CreateContestPermissionCheck(context.Context) error

	// log
	ListLogsForUser(context.Context, *ListLogsForUserRequest) (*ListLogsForUserResponse, error)
	ListLogsForContest(context.Context, *ListLogsForContestRequest) (*ListLogsForContestResponse, error)

	// profile
	ContestProfile(context.Context, *ContestProfileRequest) (*ContestProfileResponse, error)
	ActivityForContestUser(context.Context, *ActivityForContestUserRequest) (*ActivityForContestUserResponse, error)
	FetchUserProfile(context.Context, uuid.UUID) (*UserProfile, error)
	YearlyActivityForUser(context.Context, *YearlyActivityForUserRequest) (*YearlyActivityForUserResponse, error)
	YearlyScoresForUser(context.Context, *YearlyScoresForUserRequest) (*YearlyScoresForUserResponse, error)

	// admin
	ListUsers(context.Context, *ListUsersRequest) (*ListUsersResponse, error)
}

type ServiceImpl struct {
	r         Repository
	validate  *validator.Validate
	kratos    KratosClient
	clock     domain.Clock
	userCache UserCache
}

type UserCache interface {
	GetUsers() []UserEntry
}

type UserEntry struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}

type KratosClient interface {
	FetchIdentity(ctx context.Context, id uuid.UUID) (*UserTraits, error)
	ListIdentities(ctx context.Context, perPage int64, page int64) (*ListIdentitiesResult, error)
}

type ListIdentitiesResult struct {
	Identities []IdentityInfo
	HasMore    bool
}

type IdentityInfo struct {
	ID          string
	DisplayName string
	Email       string
	CreatedAt   string
}

func NewService(r Repository, clock domain.Clock, kratos KratosClient, userCache UserCache) Service {
	return &ServiceImpl{
		r:         r,
		validate:  validator.New(),
		clock:     clock,
		kratos:    kratos,
		userCache: userCache,
	}
}
