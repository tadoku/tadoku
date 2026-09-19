package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPageDelete(
	ctx context.Context,
	request openapi.ContentPageDeleteRequestObject,
) (openapi.ContentPageDeleteResponseObject, error) {
	id, err := uuid.Parse(request.Slug)
	if err != nil {
		return nil, errInvalidUUID
	}

	if err := s.application.DeletePage(ctx, request.Namespace, id); err != nil {
		s.logOperationError(ctx, "delete page", err)
		return nil, err
	}

	return openapi.ContentPageDelete204Response{}, nil
}
