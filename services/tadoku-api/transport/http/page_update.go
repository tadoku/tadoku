package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPageUpdate(
	ctx context.Context,
	request openapi.ContentPageUpdateRequestObject,
) (openapi.ContentPageUpdateResponseObject, error) {
	if request.Body == nil || request.Body.Html == nil {
		return openapi.ContentPageUpdate400Response{}, nil
	}

	id, err := uuid.Parse(request.Slug)
	if err != nil {
		return nil, errInvalidUUID
	}

	body := request.Body
	publishedAt := body.PublishedAt
	if publishedAt != nil {
		instant := publishedAt.UTC()
		publishedAt = &instant
	}
	item, err := s.application.UpdatePage(ctx, app.UpdatePageParameters{
		ID:          id,
		Namespace:   request.Namespace,
		Slug:        body.Slug,
		Title:       body.Title,
		HTML:        *body.Html,
		PublishedAt: publishedAt,
	})
	if err != nil {
		s.logOperationError(ctx, "update page", err)
		return nil, err
	}

	return openapi.ContentPageUpdate200JSONResponse{
		Id:          &item.ID,
		Slug:        item.Slug,
		Title:       item.Title,
		Html:        &item.HTML,
		PublishedAt: item.PublishedAt,
	}, nil
}
