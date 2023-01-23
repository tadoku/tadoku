package query

import (
	"context"

	"github.com/google/uuid"
)

type YearlyActivitySplitForUserRequest struct {
	UserID uuid.UUID
	Year   int
}

type YearlyActivitySplitForUserResponse struct {
	Activities []ActivityScore
}

func (s *ServiceImpl) YearlyActivitySplitForUser(ctx context.Context, req *YearlyActivitySplitForUserRequest) (*YearlyActivitySplitForUserResponse, error) {
	return s.r.YearlyActivitySplitForUser(ctx, req)
}
