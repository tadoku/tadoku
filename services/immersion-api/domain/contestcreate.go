package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type ContestCreateRepository interface {
	GetContestsByUserCountForYear(context.Context, time.Time, uuid.UUID) (int32, error)
	CreateContest(context.Context, *ContestCreateRequest) (*ContestCreateResponse, error)
}

type ContestCreateRequest struct {
	OwnerUserID             uuid.UUID `validate:"required"`
	OwnerUserDisplayName    string    `validate:"required"`
	ContestStart            time.Time `validate:"required"`
	ContestEnd              time.Time `validate:"required"`
	RegistrationEnd         time.Time `validate:"required"`
	Title                   string    `validate:"required,gt=3"`
	Description             *string
	ActivityTypeIDAllowList []int32 `validate:"required,min=1"`

	// Optional
	Official              bool
	Private               bool
	LanguageCodeAllowList []string
}

type ContestCreateResponse struct {
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
}

type ContestCreate struct {
	repo       ContestCreateRepository
	clock      commondomain.Clock
	validate   *validator.Validate
	userUpsert *UserUpsert
}

func NewContestCreate(repo ContestCreateRepository, clock commondomain.Clock, userUpsert *UserUpsert) *ContestCreate {
	return &ContestCreate{
		repo:       repo,
		clock:      clock,
		validate:   validator.New(),
		userUpsert: userUpsert,
	}
}

func (s *ContestCreate) Execute(ctx context.Context, req *ContestCreateRequest) (*ContestCreateResponse, error) {
	// Make sure the user is authorized to create a contest
	if req.Official {
		if err := requireAdmin(ctx); err != nil {
			return nil, err
		}
	} else if isGuest(ctx) {
		return nil, ErrUnauthorized
	}

	if err := s.userUpsert.Execute(ctx); err != nil {
		return nil, fmt.Errorf("could not update user: %w", err)
	}

	// Enrich request with session
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	req.OwnerUserID = uuid.MustParse(session.Subject)
	req.OwnerUserDisplayName = session.DisplayName

	// Check if user has permission to create contest
	if !isAdmin(ctx) {
		contestCount, err := s.repo.GetContestsByUserCountForYear(ctx, s.clock.Now(), req.OwnerUserID)
		if err != nil {
			return nil, fmt.Errorf("could not check permission for contest creation: %w", err)
		}

		if contestCount >= UserCreateContestYearlyLimit {
			return nil, fmt.Errorf("hit limit of created contests: %w", ErrForbidden)
		}
	}

	err := s.validate.Struct(req)
	if err != nil {
		return nil, fmt.Errorf("unable to validate: %w", ErrInvalidContest)
	}

	if req.Official && req.Private {
		return nil, fmt.Errorf("official rounds cannot be private: %w", ErrInvalidContest)
	}

	if req.Official && len(req.LanguageCodeAllowList) != 0 {
		return nil, fmt.Errorf("official rounds cannot limit language choice: %w", ErrInvalidContest)
	}

	if req.ContestStart.After(req.ContestEnd) {
		return nil, fmt.Errorf("contest cannot start after it has ended: %w", ErrInvalidContest)
	}

	if !isAdmin(ctx) {
		now := s.clock.Now()
		if now.After(req.ContestEnd) || now.After(req.ContestStart) {
			return nil, fmt.Errorf("contest cannot be in the past or already have started: %w", ErrInvalidContest)
		}
	}

	return s.repo.CreateContest(ctx, req)
}
