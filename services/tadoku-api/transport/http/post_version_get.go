package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPostVersionGet(
	ctx context.Context,
	request openapi.ContentPostVersionGetRequestObject,
) (openapi.ContentPostVersionGetResponseObject, error) {
	postID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, errInvalidUUID
	}

	version, err := s.application.GetPostVersion(ctx, request.Namespace, postID, request.ContentId)
	if err != nil {
		s.logOperationError(ctx, "get post version", err)
		return nil, err
	}

	return openapi.ContentPostVersionGet200JSONResponse{
		Id:        version.ID,
		Version:   version.Version,
		Title:     version.Title,
		Content:   &version.Content,
		CreatedAt: version.CreatedAt,
	}, nil
}
