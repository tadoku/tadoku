package query

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type Language struct {
	Code string
	Name string
}

type Activity struct {
	ID      int32
	Name    string
	Default bool
}

type FetchContestConfigurationOptionsResponse struct {
	Languages              []Language
	Activities             []Activity
	CanCreateOfficialRound bool
}

func (s *ServiceImpl) FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error) {
	res, err := s.r.FetchContestConfigurationOptions(ctx)
	if err != nil {
		return nil, err
	}

	res.CanCreateOfficialRound = domain.IsRole(ctx, domain.RoleAdmin)

	return res, nil
}

type FindContestByIDRequest struct {
	ID             uuid.UUID
	IncludeDeleted bool
}

type ContestView struct {
	ID                   uuid.UUID
	ContestStart         time.Time
	ContestEnd           time.Time
	RegistrationEnd      time.Time
	Title                string
	Description          *string
	OwnerUserID          uuid.UUID
	OwnerUserDisplayName string
	Official             bool
	Private              bool
	AllowedLanguages     []Language
	AllowedActivities    []Activity
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Deleted              bool
}

func (s *ServiceImpl) FindContestByID(ctx context.Context, req *FindContestByIDRequest) (*ContestView, error) {
	req.IncludeDeleted = domain.IsRole(ctx, domain.RoleAdmin)

	res, err := s.r.FindContestByID(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type ContestListRequest struct {
	UserID         uuid.NullUUID
	OfficialOnly   bool
	IncludeDeleted bool
	IncludePrivate bool
	PageSize       int
	Page           int
}

type ContestListResponse struct {
	Contests      []Contest
	TotalSize     int
	NextPageToken string
}

type Contest struct {
	ID                      uuid.UUID
	ContestStart            time.Time
	ContestEnd              time.Time
	RegistrationEnd         time.Time
	Title                   string
	Description             *string
	OwnerUserID             uuid.UUID
	OwnerUserDisplayName    string
	Official                bool
	Private                 bool
	LanguageCodeAllowList   []string
	ActivityTypeIDAllowList []int32
	CreatedAt               time.Time
	UpdatedAt               time.Time
	Deleted                 bool
}

func (s *ServiceImpl) ListContests(ctx context.Context, req *ContestListRequest) (*ContestListResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 || req.PageSize == 0 {
		req.PageSize = 100
	}

	req.IncludePrivate = domain.IsRole(ctx, domain.RoleAdmin)

	return s.r.ListContests(ctx, req)
}

type FindRegistrationForUserRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type ContestRegistration struct {
	ID              uuid.UUID
	ContestID       uuid.UUID
	UserID          uuid.UUID
	UserDisplayName string
	Languages       []Language
	Contest         *ContestView
}

func (s *ServiceImpl) FindRegistrationForUser(ctx context.Context, req *FindRegistrationForUserRequest) (*ContestRegistration, error) {
	session := domain.ParseSession(ctx)

	if domain.IsRole(ctx, domain.RoleGuest) || domain.IsRole(ctx, domain.RoleBanned) || session == nil {
		return nil, ErrUnauthorized
	}

	req.UserID = uuid.MustParse(session.Subject)

	res, err := s.r.FindRegistrationForUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type FetchContestLeaderboardRequest struct {
	ContestID    uuid.UUID
	LanguageCode *string
	ActivityID   *int32
	PageSize     int
	Page         int
}

type Leaderboard struct {
	Entries       []LeaderboardEntry
	TotalSize     int
	NextPageToken string
}

type LeaderboardEntry struct {
	Rank            int
	UserID          uuid.UUID
	UserDisplayName string
	Score           float32
	IsTie           bool
}

func (s *ServiceImpl) FetchContestLeaderboard(ctx context.Context, req *FetchContestLeaderboardRequest) (*Leaderboard, error) {
	if req.PageSize == 0 {
		req.PageSize = 25
	}

	if req.PageSize > 100 || req.PageSize == 0 {
		req.PageSize = 100
	}

	return s.r.FetchContestLeaderboard(ctx, req)
}

type FetchOngoingContestRegistrationsRequest struct {
	UserID uuid.UUID
	Now    time.Time
}

type ContestRegistrations struct {
	Registrations []ContestRegistration
	TotalSize     int
	NextPageToken string
}

func (s *ServiceImpl) FetchOngoingContestRegistrations(ctx context.Context, req *FetchOngoingContestRegistrationsRequest) (*ContestRegistrations, error) {
	if domain.IsRole(ctx, domain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	session := domain.ParseSession(ctx)
	req.UserID = uuid.MustParse(session.Subject)
	req.Now = s.clock.Now()

	return s.r.FetchOngoingContestRegistrations(ctx, req)
}

func (s *ServiceImpl) FindLatestOfficial(ctx context.Context) (*ContestView, error) {
	return s.r.ContestFindLatestOfficial(ctx)
}

type YearlyContestRegistrationsForUserRequest struct {
	UserID         uuid.UUID
	Year           int
	IncludePrivate bool
}

func (s *ServiceImpl) YearlyContestRegistrationsForUser(ctx context.Context, req *YearlyContestRegistrationsForUserRequest) (*ContestRegistrations, error) {
	session := domain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userId, err := uuid.Parse(session.Subject)

	sessionMatchesUser := err == nil && userId == req.UserID
	req.IncludePrivate = domain.IsRole(ctx, domain.RoleAdmin) || sessionMatchesUser

	return s.r.YearlyContestRegistrationsForUser(ctx, req)
}
