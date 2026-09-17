package http

import (
	"context"

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

	id, err := parseAnnouncementID(request.Id)
	if err != nil {
		return nil, err
	}

	body := request.Body
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
		StartsAt:  body.StartsAt,
		EndsAt:    body.EndsAt,
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "update announcement failed", "error", err)
		return nil, err
	}

	return openapi.ContentAnnouncementUpdate200JSONResponse(contentAnnouncement(item)), nil
}
