package query

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Score struct {
	LanguageCode string
	LanguageName *string
	Score        float32
}

type ContestProfileRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type ContestProfileResponse struct {
	Registration *ContestRegistration
	OverallScore float32
	Scores       []Score
}

func (s *ServiceImpl) ContestProfile(ctx context.Context, req *ContestProfileRequest) (*ContestProfileResponse, error) {

	reg, err := s.r.FindRegistrationForUser(ctx, &FindRegistrationForUserRequest{
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

func (s *ServiceImpl) ReadingActivityForContestUser(ctx context.Context, req *ContestProfileRequest) (*ReadingActivityResponse, error) {
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

func (s *ServiceImpl) FetchUserProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error) {
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

func (s *ServiceImpl) YearlyActivityForUser(ctx context.Context, req *YearlyActivityForUserRequest) (*YearlyActivityForUserResponse, error) {
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

type YearlyScoresForUserRequest struct {
	UserID uuid.UUID
	Year   int
}

type YearlyScoresForUserResponse struct {
	OverallScore float32
	Scores       []Score
}

func (s *ServiceImpl) YearlyScoresForUser(ctx context.Context, req *YearlyScoresForUserRequest) (*YearlyScoresForUserResponse, error) {
	scores, err := s.r.YearlyScoresForUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch scores: %w", err)
	}

	response := &YearlyScoresForUserResponse{
		Scores: scores,
	}

	for _, it := range scores {
		response.OverallScore += it.Score
	}

	return response, nil
}

type YearlyActivitySplitForUserRequest struct {
	UserID uuid.UUID
	Year   int
}

type ActivityScore struct {
	ActivityID   int
	ActivityName string
	Score        float32
}

type YearlyActivitySplitForUserResponse struct {
	Activities []ActivityScore
}

func (s *ServiceImpl) YearlyActivitySplitForUser(ctx context.Context, req *YearlyActivitySplitForUserRequest) (*YearlyActivitySplitForUserResponse, error) {
	return s.r.YearlyActivitySplitForUser(ctx, req)
}
