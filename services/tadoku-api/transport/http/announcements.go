package http

import (
	"context"

	"github.com/oapi-codegen/nullable"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentAnnouncementListActive(
	ctx context.Context,
	request openapi.ContentAnnouncementListActiveRequestObject,
) (openapi.ContentAnnouncementListActiveResponseObject, error) {
	items, err := s.application.ListActiveAnnouncements(ctx, request.Namespace)
	if err != nil {
		s.logger.ErrorContext(ctx, "list active announcements failed",
			"error", err,
		)
		return nil, err
	}

	response := openapi.ContentAnnouncementListActive200JSONResponse{
		Announcements: make([]openapi.ContentAnnouncement, 0, len(items)),
	}
	for _, item := range items {
		href := nullable.NewNullNullable[string]()
		if item.Href != nil {
			href = nullable.NewNullableWithValue(*item.Href)
		}

		response.Announcements = append(response.Announcements, openapi.ContentAnnouncement{
			Id:        &item.ID,
			Namespace: &item.Namespace,
			Title:     item.Title,
			Content:   item.Content,
			Style:     openapi.ContentAnnouncementStyle(item.Style),
			Href:      href,
			StartsAt:  item.StartsAt,
			EndsAt:    item.EndsAt,
			CreatedAt: &item.CreatedAt,
			UpdatedAt: &item.UpdatedAt,
		})
	}

	return response, nil
}
