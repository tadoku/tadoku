package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPageFindBySlug(
	ctx context.Context,
	request openapi.ContentPageFindBySlugRequestObject,
) (openapi.ContentPageFindBySlugResponseObject, error) {
	item, err := s.application.FindPageBySlug(ctx, request.Namespace, request.Slug)
	if err != nil {
		s.logOperationError(ctx, "find page by slug", err)
		return nil, err
	}

	return openapi.ContentPageFindBySlug200JSONResponse{
		Id:          &item.ID,
		Slug:        item.Slug,
		Title:       item.Title,
		Html:        &item.HTML,
		PublishedAt: item.PublishedAt,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}, nil
}
