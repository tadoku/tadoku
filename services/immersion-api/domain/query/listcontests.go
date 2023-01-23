package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type ListContestsRequest struct {
	UserID         uuid.NullUUID
	OfficialOnly   bool
	IncludeDeleted bool
	IncludePrivate bool
	PageSize       int
	Page           int
}

type ListContestsResponse struct {
	Contests      []Contest
	TotalSize     int
	NextPageToken string
}

func (s *ServiceImpl) ListContests(ctx context.Context, req *ListContestsRequest) (*ListContestsResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 || req.PageSize == 0 {
		req.PageSize = 100
	}

	req.IncludePrivate = domain.IsRole(ctx, domain.RoleAdmin)

	return s.r.ListContests(ctx, req)
}
