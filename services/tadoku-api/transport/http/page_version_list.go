package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPageVersionList(
	ctx context.Context,
	request openapi.ContentPageVersionListRequestObject,
) (openapi.ContentPageVersionListResponseObject, error) {
	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, errInvalidUUID
	}

	versions, err := s.application.ListPageVersions(ctx, request.Namespace, id)
	if err != nil {
		s.logOperationError(ctx, "list page versions", err)
		return nil, err
	}

	response := openapi.ContentPageVersionList200JSONResponse{
		Versions: make([]openapi.ContentPageVersion, 0, len(versions)),
	}
	for _, version := range versions {
		response.Versions = append(response.Versions, openapi.ContentPageVersion{
			Id:        version.ID,
			Version:   version.Version,
			Title:     version.Title,
			CreatedAt: version.CreatedAt,
		})
	}

	return response, nil
}
