package query

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ActivityForContestUserRequest struct {
	UserID    uuid.UUID
	ContestID uuid.UUID
}

type ActivityForContestUserResponse struct {
	Rows []ActivityForContestUserRow
}

type ActivityForContestUserRow struct {
	Date         time.Time
	LanguageCode string
	Score        float32
}

// TODO: rename input req
func (s *ServiceImpl) ActivityForContestUser(ctx context.Context, req *ActivityForContestUserRequest) (*ActivityForContestUserResponse, error) {
	rows, err := s.r.ActivityForContestUser(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch activity: %w", err)
	}

	return &ActivityForContestUserResponse{
		Rows: rows,
	}, nil
}
