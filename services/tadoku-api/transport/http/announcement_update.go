package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentAnnouncementUpdate(
	ctx context.Context,
	request openapi.ContentAnnouncementUpdateRequestObject,
) (openapi.ContentAnnouncementUpdateResponseObject, error) {
	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, errInvalidUUID
	}

	var body openapi.ContentAnnouncementUpdateJSONRequestBody
	if request.Body != nil {
		body = *request.Body
	}
	var href *string
	if value, err := body.Href.Get(); err == nil {
		href = &value
	}

	item, err := s.application.UpdateAnnouncement(ctx, app.UpdateAnnouncementParameters{
		ID:        id,
		Namespace: request.Namespace,
		Title:     body.Title,
		Content:   body.Content,
		Style:     string(body.Style),
		Href:      href,
		StartsAt:  body.StartsAt.UTC(),
		EndsAt:    body.EndsAt.UTC(),
	})
	if err != nil {
		s.logOperationError(ctx, "update announcement", err)
		return nil, err
	}

	responseHref := nullable.NewNullNullable[string]()
	if item.Href != nil {
		responseHref = nullable.NewNullableWithValue(*item.Href)
	}
	return openapi.ContentAnnouncementUpdate200JSONResponse{
		Id:        &item.ID,
		Namespace: &item.Namespace,
		Title:     item.Title,
		Content:   item.Content,
		Style:     openapi.ContentAnnouncementStyle(item.Style),
		Href:      responseHref,
		StartsAt:  item.StartsAt,
		EndsAt:    item.EndsAt,
		CreatedAt: &item.CreatedAt,
		UpdatedAt: &item.UpdatedAt,
	}, nil
}
