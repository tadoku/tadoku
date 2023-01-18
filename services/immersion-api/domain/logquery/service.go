package logquery

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

type LogRepository interface {
	ListLogsForContestUser(context.Context, *LogListForContestUserRequest) (*LogListResponse, error)
	FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error)
	FindLogByID(context.Context, *FindLogByIDRequest) (*Log, error)
}

type Service interface {
	ListLogsForContestUser(context.Context, *LogListForContestUserRequest) (*LogListResponse, error)
	FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error)
	FindLogByID(context.Context, *FindLogByIDRequest) (*Log, error)
}

type service struct {
	r        LogRepository
	validate *validator.Validate
	clock    domain.Clock
}

func NewService(r LogRepository, clock domain.Clock) Service {
	return &service{
		r:        r,
		validate: validator.New(),
		clock:    clock,
	}
}

type Language struct {
	Code string
	Name string
}

type Activity struct {
	ID      int32
	Name    string
	Default bool
}

type Unit struct {
	ID            uuid.UUID
	LogActivityID int
	Name          string
	Modifier      float32
	LanguageCode  *string
}

type Tag struct {
	ID            uuid.UUID
	LogActivityID int
	Name          string
}

type FetchLogConfigurationOptionsResponse struct {
	Languages  []Language
	Activities []Activity
	Units      []Unit
	Tags       []Tag
}

func (s *service) FetchLogConfigurationOptions(ctx context.Context) (*FetchLogConfigurationOptionsResponse, error) {
	if domain.IsRole(ctx, domain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	return s.r.FetchLogConfigurationOptions(ctx)
}

type LogListForContestUserRequest struct {
	UserID         uuid.UUID
	ContestID      uuid.UUID
	IncludeDeleted bool
	PageSize       int
	Page           int
}

type ContestRegistrationReference struct {
	RegistrationID uuid.UUID
	ContestID      uuid.UUID
	Title          string
}

type Log struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	UserDisplayName *string
	Description     *string
	LanguageCode    string
	LanguageName    string
	ActivityID      int
	ActivityName    string
	UnitName        string
	Tags            []string
	Amount          float32
	Modifier        float32
	Score           float32
	Registrations   []ContestRegistrationReference
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Deleted         bool
}

type LogListResponse struct {
	Logs          []Log
	TotalSize     int
	NextPageToken string
}

func (s *service) ListLogsForContestUser(ctx context.Context, req *LogListForContestUserRequest) (*LogListResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 50
	}

	if req.PageSize > 100 || req.PageSize < 0 {
		req.PageSize = 100
	}

	if req.IncludeDeleted && !domain.IsRole(ctx, domain.RoleAdmin) {
		return nil, ErrUnauthorized
	}

	return s.r.ListLogsForContestUser(ctx, req)
}

type FindLogByIDRequest struct {
	ID             uuid.UUID
	IncludeDeleted bool
}

func (s *service) FindLogByID(ctx context.Context, req *FindLogByIDRequest) (*Log, error) {
	req.IncludeDeleted = domain.IsRole(ctx, domain.RoleAdmin)

	log, err := s.r.FindLogByID(ctx, req)
	if err != nil {
		return nil, err
	}

	session := domain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	userID := uuid.MustParse(session.Subject)

	// Needed to prevent leaking private registrations, only show to admins and the owner of the log
	if !domain.IsRole(ctx, domain.RoleAdmin) && log.UserID != userID {
		log.Registrations = nil
	}

	return log, nil
}
