package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPostVersionList(
	ctx context.Context,
	request openapi.ContentPostVersionListRequestObject,
) (openapi.ContentPostVersionListResponseObject, error) {
	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, errInvalidUUID
	}

	versions, err := s.application.ListPostVersions(ctx, request.Namespace, id)
	if err != nil {
		s.logOperationError(ctx, "list post versions", err)
		return nil, err
	}

	response := openapi.ContentPostVersionList200JSONResponse{
		Versions: make([]openapi.ContentPostVersion, 0, len(versions)),
	}
	for _, version := range versions {
		response.Versions = append(response.Versions, openapi.ContentPostVersion{
			Id:        version.ID,
			Version:   version.Version,
			Title:     version.Title,
			CreatedAt: version.CreatedAt,
		})
	}

	return response, nil
}
