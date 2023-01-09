package logcommand

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logquery"
)

var ErrInvalidLog = errors.New("unable to validate log")
var ErrForbidden = errors.New("not allowed")
var ErrUnauthorized = errors.New("unauthorized")

type LogRepository interface {
	CreateLog(context.Context, *LogCreateRequest) error

	logquery.LogRepository
}

type Service interface {
	CreateLog(context.Context, *LogCreateRequest) error
}

type service struct {
	lr       LogRepository
	cr       contestquery.ContestRepository
	validate *validator.Validate
	clock    domain.Clock
}

func NewService(
	lr LogRepository,
	cr contestquery.ContestRepository,
	clock domain.Clock,
) Service {
	return &service{
		lr:       lr,
		cr:       cr,
		validate: validator.New(),
		clock:    clock,
	}
}

type LogCreateRequest struct {
	RegistrationIDs []uuid.UUID `validate:"required"`
	UnitID          uuid.UUID   `validate:"required"`
	UserID          uuid.UUID   `validate:"required"`
	ActivityID      int32       `validate:"required"`
	LanguageCode    string      `validate:"required"`
	Amount          float32     `validate:"required,gte=0"`
	Tags            []string

	// Optional
	Description                 *string
	EligibleOfficialLeaderboard bool
}

func (s *service) CreateLog(ctx context.Context, req *LogCreateRequest) error {
	// Make sure the user is authorized to create a contest
	if domain.IsRole(ctx, domain.RoleGuest) {
		return ErrUnauthorized
	}

	// Enrich request with session
	session := domain.ParseSession(ctx)
	if session == nil {
		return ErrUnauthorized
	}
	req.UserID = uuid.MustParse(session.Subject)

	err := s.validate.Struct(req)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("unable to validate: %w", ErrInvalidLog)
	}

	registrations, err := s.cr.FetchOngoingContestRegistrations(ctx, &contestquery.FetchOngoingContestRegistrationsRequest{
		UserID: req.UserID,
		Now:    s.clock.Now(),
	})
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("unable to fetch registrations: %w", err)
	}

	validContestIDs := map[uuid.UUID]contestquery.ContestRegistration{}
	for _, r := range registrations.Registrations {
		validContestIDs[r.ID] = r
	}

	// validate registrations
	for _, id := range req.RegistrationIDs {
		registration, ok := validContestIDs[id]
		if !ok {
			return fmt.Errorf("registration is not found as ongoing for the current user: %w", ErrInvalidLog)
		}

		if registration.Contest.Official {
			req.EligibleOfficialLeaderboard = true
		}

		// validate language is part of registration
		found := false
		for _, lang := range registration.Languages {
			if lang.Code == req.LanguageCode {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("language is not allowed for registration: %w", ErrInvalidLog)
		}

		// validate activity is allowed by the contest
		found = false
		for _, act := range registration.Contest.AllowedActivities {
			if act.ID == req.ActivityID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("activity is not allowed for registration: %w", ErrInvalidLog)
		}
	}

	return s.lr.CreateLog(ctx, req)
}
