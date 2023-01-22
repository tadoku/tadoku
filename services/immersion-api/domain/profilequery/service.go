package profilequery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain/contestquery"
)

var ErrRequestInvalid = errors.New("request is invalid")
var ErrNotFound = errors.New("not found")
var ErrUnauthorized = errors.New("unauthorized")

type Repository interface {
	FindRegistrationForUser(context.Context, *contestquery.FindRegistrationForUserRequest) (*contestquery.ContestRegistration, error)
	FindScoresForRegistration(context.Context, *ContestProfileRequest) ([]Score, error)
	ReadingActivityForContestUser(context.Context, *ContestProfileRequest) ([]ReadingActivityRow, error)
}

type Service interface {
	ContestProfile(context.Context, *ContestProfileRequest) (*ContestProfileResponse, error)
	ReadingActivityForContestUser(context.Context, *ContestProfileRequest) (*ReadingActivityResponse, error)
}

type service struct {
	r        Repository
	validate *validator.Validate
}

func NewService(r Repository) Service {
	return &service{
		r:        r,
		validate: validator.New(),
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

	scores, err := s.r.FindScoresForRegistration(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	response.Scores = scores

	for _, it := range scores {
		response.OverallScore += it.Score
	}

	return response, nil
}

type ReadingActivityResponse struct {
	Rows []ReadingActivityRow
}

type ReadingActivityRow struct {
	Date         time.Time
	LanguageCode string
	Score        float32
}

func (s *service) ReadingActivityForContestUser(ctx context.Context, req *ContestProfileRequest) (*ReadingActivityResponse, error) {
	rows, err := s.r.ReadingActivityForContestUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch reading activity: %w", err)
	}

	return &ReadingActivityResponse{
		Rows: rows,
	}, nil
}

type UserTraits struct {
	UserDisplayName string
	Email           string
}
