package http

import (
	"context"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/app/api"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/uuidx"
)

func (s *server) ContentPostCreate(
	ctx context.Context,
	request openapi.ContentPostCreateRequestObject,
) (openapi.ContentPostCreateResponseObject, error) {
	var body openapi.ContentPostCreateJSONRequestBody
	if request.Body != nil {
		body = *request.Body
	}
	id := uuidx.New(ctx)
	if body.Id != nil {
		id = *body.Id
	}
	var publishedAt *time.Time
	if body.PublishedAt != nil {
		instant := body.PublishedAt.UTC()
		publishedAt = &instant
	}

	item, err := s.application.CreatePost(ctx, api.CreatePostParameters{
		ID:          id,
		Namespace:   request.Namespace,
		Slug:        body.Slug,
		Title:       body.Title,
		Content:     body.Content,
		PublishedAt: publishedAt,
	})
	if err != nil {
		s.logOperationError(ctx, "create post", err)
		return nil, err
	}

	return openapi.ContentPostCreate201JSONResponse{
		Id:          &item.ID,
		Slug:        item.Slug,
		Title:       item.Title,
		Content:     item.Content,
		PublishedAt: item.PublishedAt,
	}, nil
}
