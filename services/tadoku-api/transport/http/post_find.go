package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPostFindBySlug(
	ctx context.Context,
	request openapi.ContentPostFindBySlugRequestObject,
) (openapi.ContentPostFindBySlugResponseObject, error) {
	item, err := s.application.FindPostBySlug(ctx, request.Namespace, request.Slug)
	if err != nil {
		s.logOperationError(ctx, "find post by slug", err)
		return nil, err
	}

	return openapi.ContentPostFindBySlug200JSONResponse{
		Id:          &item.ID,
		Slug:        item.Slug,
		Title:       item.Title,
		Content:     item.Content,
		PublishedAt: item.PublishedAt,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}, nil
}
