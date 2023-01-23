package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type FindContestByIDRequest struct {
	ID             uuid.UUID
	IncludeDeleted bool
}

func (s *ServiceImpl) FindContestByID(ctx context.Context, req *FindContestByIDRequest) (*ContestView, error) {
	req.IncludeDeleted = domain.IsRole(ctx, domain.RoleAdmin)

	res, err := s.r.FindContestByID(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
