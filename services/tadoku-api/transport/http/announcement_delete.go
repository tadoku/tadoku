package http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

// The legacy spec omits this response, so preserve its empty 400 locally.
type announcementDeleteBadRequestResponse struct{}

func (announcementDeleteBadRequestResponse) VisitContentAnnouncementDeleteResponse(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusBadRequest)
	return nil
}

func (s *server) ContentAnnouncementDelete(
	ctx context.Context,
	request openapi.ContentAnnouncementDeleteRequestObject,
) (openapi.ContentAnnouncementDeleteResponseObject, error) {
	id, err := uuid.Parse(request.Id)
	if err != nil {
		return announcementDeleteBadRequestResponse{}, nil
	}

	if err := s.application.DeleteAnnouncement(ctx, request.Namespace, id); err != nil {
		s.logger.ErrorContext(ctx, "delete announcement failed",
			"error", err,
		)
		return nil, err
	}

	return openapi.ContentAnnouncementDelete204Response{}, nil
}
