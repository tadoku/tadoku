package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentAnnouncementDelete(
	ctx context.Context,
	request openapi.ContentAnnouncementDeleteRequestObject,
) (openapi.ContentAnnouncementDeleteResponseObject, error) {
	id, err := parseAnnouncementID(request.Id)
	if err != nil {
		return nil, err
	}

	if err := s.application.DeleteAnnouncement(ctx, request.Namespace, id); err != nil {
		s.logger.ErrorContext(ctx, "delete announcement failed",
			"error", err,
		)
		return nil, err
	}

	return openapi.ContentAnnouncementDelete204Response{}, nil
}
