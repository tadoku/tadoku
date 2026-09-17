package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentAnnouncementDelete(
	ctx context.Context,
	request openapi.ContentAnnouncementDeleteRequestObject,
) (openapi.ContentAnnouncementDeleteResponseObject, error) {
	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, errInvalidUUID
	}

	if err := s.application.DeleteAnnouncement(ctx, request.Namespace, id); err != nil {
		s.logOperationError(ctx, "delete announcement", err)
		return nil, err
	}

	return openapi.ContentAnnouncementDelete204Response{}, nil
}
