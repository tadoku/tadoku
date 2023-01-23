package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
)

var ErrInvalidContest = errors.New("unable to validate contest")
var ErrInvalidContestRegistration = errors.New("language selection is not valid for contest")

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

func (s *ServiceImpl) CreateContest(ctx context.Context, req *ContestCreateRequest) (*ContestCreateResponse, error) {
	// Make sure the user is authorized to create a contest
	if domain.IsRole(ctx, domain.RoleBanned) {
		return nil, ErrForbidden
	}
	if domain.IsRole(ctx, domain.RoleGuest) {
		return nil, ErrUnauthorized
	}
	if req.Official && !domain.IsRole(ctx, domain.RoleAdmin) {
		return nil, ErrForbidden
	}

	// Enrich request with session
	session := domain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}
	req.OwnerUserID = uuid.MustParse(session.Subject)
	req.OwnerUserDisplayName = session.DisplayName

	err := s.validate.Struct(req)
	if err != nil {
		fmt.Println(err)
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

	if !domain.IsRole(ctx, domain.RoleAdmin) {
		now := s.clock.Now()
		if now.After(req.ContestEnd) || now.After(req.ContestStart) {
			return nil, fmt.Errorf("contest cannot be in the past or already have started: %w", ErrInvalidContest)
		}
	}

	return s.r.CreateContest(ctx, req)
}

type UpsertContestRegistrationRequest struct {
	ID              uuid.UUID
	ContestID       uuid.UUID
	UserID          uuid.UUID
	UserDisplayName string
	LanguageCodes   []string
}

func (s *ServiceImpl) UpsertContestRegistration(ctx context.Context, req *UpsertContestRegistrationRequest) error {
	if domain.IsRole(ctx, domain.RoleBanned) {
		return ErrForbidden
	}
	if domain.IsRole(ctx, domain.RoleGuest) {
		return ErrUnauthorized
	}

	// Enrich request with session
	session := domain.ParseSession(ctx)
	if session == nil {
		return ErrUnauthorized
	}
	req.UserID = uuid.MustParse(session.Subject)
	req.UserDisplayName = session.DisplayName
	req.ID = uuid.New()

	// TODO: should rename to FindContestByID
	contest, err := s.r.FindByID(ctx, &query.FindByIDRequest{
		ID:             req.ContestID,
		IncludeDeleted: false,
	})
	if err != nil {
		return fmt.Errorf("could not find contest: %w", err)
	}

	if len(req.LanguageCodes) < 1 || len(req.LanguageCodes) > 3 {
		return fmt.Errorf("invalid language code length: %w", ErrInvalidContestRegistration)
	}

	// check if languages are allowed by contest
	if len(contest.AllowedLanguages) > 0 {
		langs := map[string]bool{}
		for _, lang := range contest.AllowedLanguages {
			langs[lang.Code] = true
		}
		for _, code := range req.LanguageCodes {
			if _, ok := langs[code]; !ok {
				return fmt.Errorf("language %s is not allowed by contest: %w", code, ErrInvalidContestRegistration)
			}
		}
	}

	// check if existing registration
	registration, err := s.r.FindRegistrationForUser(ctx, &query.FindRegistrationForUserRequest{
		UserID:    req.UserID,
		ContestID: req.ContestID,
	})
	if err != nil && !errors.Is(err, query.ErrNotFound) {
		return err
	}

	// check if previous languages are included in new set
	if registration != nil {
		req.ID = registration.ID

		langs := map[string]bool{}
		for _, lang := range req.LanguageCodes {
			langs[lang] = true
		}
		for _, lang := range registration.Languages {
			if _, ok := langs[lang.Code]; !ok {
				return fmt.Errorf("language %s is missing but was previously registered: %w", lang.Code, ErrInvalidContestRegistration)
			}
		}
	}

	return s.r.UpsertContestRegistration(ctx, req)
}
