package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPostDelete(
	ctx context.Context,
	request openapi.ContentPostDeleteRequestObject,
) (openapi.ContentPostDeleteResponseObject, error) {
	id, err := uuid.Parse(request.Slug)
	if err != nil {
		return nil, errInvalidUUID
	}

	if err := s.application.DeletePost(ctx, request.Namespace, id); err != nil {
		s.logOperationError(ctx, "delete post", err)
		return nil, err
	}

	return openapi.ContentPostDelete204Response{}, nil
}
