package logcommand

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/logquery"
)

var ErrInvalidLog = errors.New("unable to validate log")
var ErrForbidden = errors.New("not allowed")
var ErrUnauthorized = errors.New("unauthorized")

type ContestRepository interface {
	CreateLog(context.Context, *LogCreateRequest) error

	logquery.LogRepository
}

type Service interface {
	CreateLog(context.Context, *LogCreateRequest) error
}

type service struct {
	r        ContestRepository
	validate *validator.Validate
	clock    domain.Clock
}

func NewService(r ContestRepository, clock domain.Clock) Service {
	return &service{
		r:        r,
		validate: validator.New(),
		clock:    clock,
	}
}

type LogCreateRequest struct {
	RegistrationIDs []uuid.UUID `validate:"required"`
	ActivityID      uuid.UUID   `validate:"required"`
	UnitID          uuid.UUID   `validate:"required"`
	UserID          uuid.UUID   `validate:"required"`
	LanguageCode    string      `validate:"required"`
	Amount          float32     `validate:"required,gte=0"`
	Tags            []string

	// Optional
	Description *string
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

	return s.r.CreateLog(ctx, req)
}
