package query

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	immersiondomain "github.com/tadoku/tadoku/services/immersion-api/domain"
)

type ContestProfileRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type ContestProfileResponse struct {
	Registration *immersiondomain.ContestRegistration
	OverallScore float32
	Scores       []Score
}

func (s *ServiceImpl) ContestProfile(ctx context.Context, req *ContestProfileRequest) (*ContestProfileResponse, error) {

	reg, err := s.r.FindRegistrationForUser(ctx, &immersiondomain.RegistrationFindRequest{
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
