package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentAnnouncementCreate(
	ctx context.Context,
	request openapi.ContentAnnouncementCreateRequestObject,
) (openapi.ContentAnnouncementCreateResponseObject, error) {
	var body openapi.ContentAnnouncementCreateJSONRequestBody
	if request.Body != nil {
		body = *request.Body
	}
	id := uuid.New()
	if body.Id != nil {
		id = *body.Id
	}
	var href *string
	if value, err := body.Href.Get(); err == nil {
		href = &value
	}

	item, err := s.application.CreateAnnouncement(ctx, app.CreateAnnouncementParameters{
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
		s.logOperationError(ctx, "create announcement", err)
		return nil, err
	}

	return openapi.ContentAnnouncementCreate201JSONResponse(announcementResponse(item)), nil
}
