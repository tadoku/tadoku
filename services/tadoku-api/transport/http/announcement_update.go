package http

import (
	"context"

	"github.com/google/uuid"
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

	return openapi.ContentAnnouncementUpdate200JSONResponse(announcementResponse(item)), nil
}
