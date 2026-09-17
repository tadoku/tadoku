package http

import (
	"context"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentAnnouncementFindByID(
	ctx context.Context,
	request openapi.ContentAnnouncementFindByIDRequestObject,
) (openapi.ContentAnnouncementFindByIDResponseObject, error) {
	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, errInvalidUUID
	}

	item, err := s.application.FindAnnouncementByID(ctx, request.Namespace, id)
	if err != nil {
		s.logOperationError(ctx, "find announcement by ID", err)
		return nil, err
	}

	href := nullable.NewNullNullable[string]()
	if item.Href != nil {
		href = nullable.NewNullableWithValue(*item.Href)
	}

	return openapi.ContentAnnouncementFindByID200JSONResponse{
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
	}, nil
}

func (s *server) ContentAnnouncementListActive(
	ctx context.Context,
	request openapi.ContentAnnouncementListActiveRequestObject,
) (openapi.ContentAnnouncementListActiveResponseObject, error) {
	items, err := s.application.ListActiveAnnouncements(ctx, request.Namespace)
	if err != nil {
		s.logOperationError(ctx, "list active announcements", err)
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
		s.logOperationError(ctx, "list announcements", err)
		return nil, err
	}

	response := openapi.ContentAnnouncementList200JSONResponse{
		Announcements: make([]openapi.ContentAnnouncement, 0, len(result.Announcements)),
		NextPageToken: result.NextPageToken,
		TotalSize:     result.TotalSize,
	}
	for _, item := range result.Announcements {
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
