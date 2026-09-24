package http

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPageCreate(
	ctx context.Context,
	request openapi.ContentPageCreateRequestObject,
) (openapi.ContentPageCreateResponseObject, error) {
	var body openapi.ContentPageCreateJSONRequestBody
	if request.Body != nil {
		body = *request.Body
	}
	id := uuid.New()
	if body.Id != nil {
		id = *body.Id
	}
	var publishedAt *time.Time
	if body.PublishedAt != nil {
		instant := body.PublishedAt.UTC()
		publishedAt = &instant
	}
	var html string
	if body.Html != nil {
		html = *body.Html
	}

	item, err := s.application.CreatePage(ctx, app.CreatePageParameters{
		ID:          id,
		Namespace:   request.Namespace,
		Slug:        body.Slug,
		Title:       body.Title,
		HTML:        html,
		PublishedAt: publishedAt,
	})
	if err != nil {
		s.logOperationError(ctx, "create page", err)
		return nil, err
	}

	return openapi.ContentPageCreate201JSONResponse{
		Id:          &item.ID,
		Slug:        item.Slug,
		Title:       item.Title,
		Html:        &item.HTML,
		PublishedAt: item.PublishedAt,
	}, nil
}
