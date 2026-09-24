package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPostUpdate(
	ctx context.Context,
	request openapi.ContentPostUpdateRequestObject,
) (openapi.ContentPostUpdateResponseObject, error) {
	id, err := uuid.Parse(request.Slug)
	if err != nil {
		return nil, errInvalidUUID
	}

	var body openapi.ContentPostUpdateJSONRequestBody
	if request.Body != nil {
		body = *request.Body
	}
	publishedAt := body.PublishedAt
	if publishedAt != nil {
		value := publishedAt.UTC()
		publishedAt = &value
	}
	item, err := s.application.UpdatePost(ctx, app.UpdatePostParameters{
		ID:          id,
		Namespace:   request.Namespace,
		Slug:        body.Slug,
		Title:       body.Title,
		Content:     body.Content,
		PublishedAt: publishedAt,
	})
	if err != nil {
		s.logOperationError(ctx, "update post", err)
		return nil, err
	}

	return openapi.ContentPostUpdate200JSONResponse{
		Id:          &item.ID,
		Slug:        item.Slug,
		Title:       item.Title,
		Content:     item.Content,
		PublishedAt: item.PublishedAt,
	}, nil
}
