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
	YearlyActivityForUser(context.Context, *YearlyActivityForUserRequest) ([]UserActivityScore, error)
}

type Service interface {
	ContestProfile(context.Context, *ContestProfileRequest) (*ContestProfileResponse, error)
	// TODO: Shouldn't include reading prefix
	ReadingActivityForContestUser(context.Context, *ContestProfileRequest) (*ReadingActivityResponse, error)
	FetchUserProfile(context.Context, uuid.UUID) (*UserProfile, error)
	YearlyActivityForUser(context.Context, *YearlyActivityForUserRequest) (*YearlyActivityForUserResponse, error)
}

type KratosClient interface {
	FetchIdentity(ctx context.Context, id uuid.UUID) (*UserTraits, error)
}

type service struct {
	r        Repository
	kratos   KratosClient
	validate *validator.Validate
}

func NewService(r Repository, kratos KratosClient) Service {
	return &service{
		r:        r,
		validate: validator.New(),
		kratos:   kratos,
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
	CreatedAt       time.Time
}

type UserProfile struct {
	DisplayName string
	CreatedAt   time.Time
}

func (s *service) FetchUserProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error) {
	traits, err := s.kratos.FetchIdentity(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not fetch user profile: %w", err)
	}

	return &UserProfile{
		DisplayName: traits.UserDisplayName,
		CreatedAt:   traits.CreatedAt,
	}, nil
}

type UserActivityScore struct {
	Date    time.Time
	Score   float32
	Updates int
}

type YearlyActivityForUserRequest struct {
	UserID uuid.UUID
	Year   int
}

type YearlyActivityForUserResponse struct {
	Scores       []UserActivityScore
	TotalUpdates int
}

func (s *service) YearlyActivityForUser(ctx context.Context, req *YearlyActivityForUserRequest) (*YearlyActivityForUserResponse, error) {
	scores, err := s.r.YearlyActivityForUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch activity summary: %w", err)
	}

	res := &YearlyActivityForUserResponse{
		Scores:       scores,
		TotalUpdates: 0,
	}

	for _, it := range scores {
		res.TotalUpdates += it.Updates
	}

	return res, nil
}
