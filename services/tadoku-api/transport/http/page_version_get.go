package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPageVersionGet(
	ctx context.Context,
	request openapi.ContentPageVersionGetRequestObject,
) (openapi.ContentPageVersionGetResponseObject, error) {
	pageID, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, errInvalidUUID
	}

	version, err := s.application.GetPageVersion(ctx, request.Namespace, pageID, request.ContentId)
	if err != nil {
		s.logOperationError(ctx, "get page version", err)
		return nil, err
	}

	return openapi.ContentPageVersionGet200JSONResponse{
		Id:        version.ID,
		Version:   version.Version,
		Title:     version.Title,
		Html:      &version.HTML,
		CreatedAt: version.CreatedAt,
	}, nil
}
