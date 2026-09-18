package http

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPostCreate(
	ctx context.Context,
	request openapi.ContentPostCreateRequestObject,
) (openapi.ContentPostCreateResponseObject, error) {
	if request.Body == nil {
		return openapi.ContentPostCreate400Response{}, nil
	}
	body := request.Body
	id := uuid.New()
	if body.Id != nil {
		id = *body.Id
	}
	var publishedAt *time.Time
	if body.PublishedAt != nil {
		instant := body.PublishedAt.UTC()
		publishedAt = &instant
	}

	item, err := s.application.CreatePost(ctx, app.CreatePostParameters{
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
