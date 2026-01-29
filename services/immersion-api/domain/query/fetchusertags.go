package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
)

type UserTag struct {
	Tag        string
	UsageCount int64
}

type FetchUserTagsRequest struct {
	UserID uuid.UUID
	Prefix string
	Limit  int
	Offset int
}

type FetchUserTagsResponse struct {
	Tags        []UserTag
	DefaultTags []string
}

func (s *ServiceImpl) FetchUserTags(ctx context.Context, req *FetchUserTagsRequest) (*FetchUserTagsResponse, error) {
	if domain.IsRole(ctx, domain.RoleGuest) {
		return nil, ErrUnauthorized
	}

	return s.r.FetchUserTags(ctx, req)
}
