package profilequery

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
)

var ErrRequestInvalid = errors.New("request is invalid")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

type Repository interface {
	FindRegistrationForUser(context.Context, *contestquery.FindRegistrationForUserRequest) (*contestquery.ContestRegistration, error)
	FindScoresForRegistration(context.Context, *ContestProfileRequest) ([]Score, error)
}

type Service interface {
	ContestProfile(context.Context, *ContestProfileRequest) (*ContestProfileResponse, error)
}

type service struct {
	r        Repository
	validate *validator.Validate
	clock    domain.Clock
}

func NewService(r Repository, clock domain.Clock) Service {
	return &service{
		r:        r,
		validate: validator.New(),
		clock:    clock,
	}
}

type Score struct {
	LanguageCode string
	Score        float32
}

type ContestProfileRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type ContestProfileResponse struct {
	Registration *contestquery.ContestRegistration
	OverallScore float32
	Scores       []Score
}

func (s *service) ContestProfile(ctx context.Context, req *ContestProfileRequest) (*ContestProfileResponse, error) {

	reg, err := s.r.FindRegistrationForUser(ctx, &contestquery.FindRegistrationForUserRequest{
		UserID:    req.UserID,
		ContestID: req.ContestID,
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch registration: %w", err)
	}

	response := &ContestProfileResponse{
		Registration: reg,
	}

	return response, nil
}
