package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func parseAnnouncementID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, errInvalidUUID
	}
	return id, nil
}

func contentAnnouncement(item *app.Announcement) openapi.ContentAnnouncement {
	href := nullable.NewNullNullable[string]()
	if item.Href != nil {
		href = nullable.NewNullableWithValue(*item.Href)
	}

	return openapi.ContentAnnouncement{
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
	}
}

func (s *server) ContentAnnouncementFindByID(
	ctx context.Context,
	request openapi.ContentAnnouncementFindByIDRequestObject,
) (openapi.ContentAnnouncementFindByIDResponseObject, error) {
	id, err := parseAnnouncementID(request.Id)
	if err != nil {
		return nil, err
	}

	item, err := s.application.FindAnnouncementByID(ctx, request.Namespace, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "find announcement by ID failed",
			"error", err,
		)
		return nil, err
	}

	return openapi.ContentAnnouncementFindByID200JSONResponse(contentAnnouncement(item)), nil
}

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
	for i := range items {
		response.Announcements = append(response.Announcements, contentAnnouncement(&items[i]))
	}

	return response, nil
}

func (s *server) ContentAnnouncementList(
	ctx context.Context,
	request openapi.ContentAnnouncementListRequestObject,
) (openapi.ContentAnnouncementListResponseObject, error) {
	pageSize, page := 0, 0
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}
	if request.Params.Page != nil {
		page = *request.Params.Page
	}

	result, err := s.application.ListAnnouncements(ctx, request.Namespace, pageSize, page)
	if err != nil {
		s.logger.ErrorContext(ctx, "list announcements failed",
			"error", err,
		)
		return nil, err
	}

	response := openapi.ContentAnnouncementList200JSONResponse{
		Announcements: make([]openapi.ContentAnnouncement, 0, len(result.Announcements)),
		NextPageToken: result.NextPageToken,
		TotalSize:     result.TotalSize,
	}
	for i := range result.Announcements {
		response.Announcements = append(response.Announcements, contentAnnouncement(&result.Announcements[i]))
	}

	return response, nil
}
