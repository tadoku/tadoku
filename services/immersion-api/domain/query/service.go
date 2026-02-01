package query

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

type Repository interface {
	// contest
	FetchContestConfigurationOptions(ctx context.Context) (*domain.ContestConfigurationOptionsResponse, error)
	FindContestByID(context.Context, *domain.ContestFindRequest) (*domain.ContestView, error)
	ContestFindLatestOfficial(context.Context) (*domain.ContestView, error)
	ListContests(context.Context, *domain.ContestListRequest) (*domain.ContestListResponse, error)
	FindRegistrationForUser(context.Context, *domain.RegistrationFindRequest) (*domain.ContestRegistration, error)
	FetchContestLeaderboard(context.Context, *domain.ContestLeaderboardFetchRequest) (*domain.Leaderboard, error)
	YearlyContestRegistrationsForUser(context.Context, *domain.RegistrationListYearlyRequest) (*domain.ContestRegistrations, error)
	FetchYearlyLeaderboard(context.Context, *domain.LeaderboardYearlyRequest) (*domain.Leaderboard, error)
	FetchGlobalLeaderboard(context.Context, *domain.LeaderboardGlobalRequest) (*domain.Leaderboard, error)
	FetchContestSummary(context.Context, *domain.ContestSummaryFetchRequest) (*domain.ContestSummaryFetchResponse, error)
	GetContestsByUserCountForYear(context.Context, time.Time, uuid.UUID) (int32, error)

	// log
	ListLogsForUser(context.Context, *domain.LogListForUserRequest) (*domain.LogListForUserResponse, error)
	ListLogsForContest(context.Context, *domain.LogListForContestRequest) (*domain.LogListForContestResponse, error)
	FetchLogConfigurationOptions(ctx context.Context) (*domain.LogConfigurationOptionsResponse, error)
	FindLogByID(context.Context, *domain.LogFindRequest) (*domain.Log, error)

	// profile
	FindScoresForRegistration(context.Context, *domain.ProfileContestRequest) ([]domain.Score, error)
	ActivityForContestUser(context.Context, *domain.ProfileContestActivityRequest) ([]domain.ProfileContestActivityRow, error)
	YearlyActivityForUser(context.Context, *domain.ProfileYearlyActivityRequest) ([]domain.UserActivityScore, error)
	YearlyScoresForUser(context.Context, *domain.ProfileYearlyScoresRequest) ([]domain.Score, error)
	YearlyActivitySplitForUser(context.Context, *domain.ProfileYearlyActivitySplitRequest) (*domain.ProfileYearlyActivitySplitResponse, error)
}
