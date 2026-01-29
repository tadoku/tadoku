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
	session := domain.ParseSession(ctx)
	if session == nil {
		return nil, ErrUnauthorized
	}

	sessionUserID, err := uuid.Parse(session.Subject)
	if err != nil || sessionUserID != req.UserID {
		return nil, ErrUnauthorized
	}

	return s.r.FetchUserTags(ctx, req)
}
