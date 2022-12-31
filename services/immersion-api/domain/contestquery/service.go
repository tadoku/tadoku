package contestquery

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

var ErrRequestInvalid = errors.New("request is invalid")

type ContestRepository interface {
	FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error)
	ListContests(context.Context, *ContestListRequest) (*ContestListResponse, error)
}

type Service interface {
	FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error)
	ListContests(context.Context, *ContestListRequest) (*ContestListResponse, error)
}

type service struct {
	r        ContestRepository
	validate *validator.Validate
}

func NewService(r ContestRepository) Service {
	return &service{
		r:        r,
		validate: validator.New(),
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

type FetchContestConfigurationOptionsResponse struct {
	Languages              []Language
	Activities             []Activity
	CanCreateOfficialRound bool
}

func (s *service) FetchContestConfigurationOptions(ctx context.Context) (*FetchContestConfigurationOptionsResponse, error) {
	res, err := s.r.FetchContestConfigurationOptions(ctx)
	if err != nil {
		return nil, err
	}

	res.CanCreateOfficialRound = domain.IsRole(ctx, domain.RoleAdmin)

	return res, nil
}

type ContestListRequest struct {
	UserID         uuid.NullUUID
	OfficialOnly   bool
	IncludeDeleted bool
	PageSize       int
	Page           int
}

type ContestListResponse struct {
	Contests      []ContestListEntry
	TotalSize     int
	NextPageToken string
}

type ContestListEntry struct {
	ID                      uuid.UUID
	ContestStart            time.Time
	ContestEnd              time.Time
	RegistrationStart       time.Time
	RegistrationEnd         time.Time
	Description             string
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

func (s *service) ListContests(ctx context.Context, req *ContestListRequest) (*ContestListResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 || req.PageSize == 0 {
		req.PageSize = 100
	}

	return s.r.ListContests(ctx, req)
}
