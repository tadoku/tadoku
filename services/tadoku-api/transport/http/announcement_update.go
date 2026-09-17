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
	if request.Body == nil {
		return openapi.ContentAnnouncementUpdate400Response{}, nil
	}

	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, errInvalidUUID
	}

	body := request.Body
	var href *string
	if body.Href.IsSpecified() && !body.Href.IsNull() {
		value := body.Href.MustGet()
		href = &value
	}

	item, err := s.application.UpdateAnnouncement(ctx, request.Namespace, id, app.UpdateAnnouncementParameters{
		Title:    body.Title,
		Content:  body.Content,
		Style:    string(body.Style),
		Href:     href,
		StartsAt: body.StartsAt,
		EndsAt:   body.EndsAt,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "update announcement failed", "error", err)
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
