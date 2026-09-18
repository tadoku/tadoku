package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPageList(
	ctx context.Context,
	request openapi.ContentPageListRequestObject,
) (openapi.ContentPageListResponseObject, error) {
	pageSize, page := 0, 0
	includeDrafts := true
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}
	if request.Params.Page != nil {
		page = *request.Params.Page
	}
	if request.Params.IncludeDrafts != nil {
		includeDrafts = *request.Params.IncludeDrafts
	}

	result, err := s.application.ListPages(ctx, request.Namespace, includeDrafts, pageSize, page)
	if err != nil {
		s.logOperationError(ctx, "list pages", err)
		return nil, err
	}

	response := openapi.ContentPageList200JSONResponse{
		Pages:         make([]openapi.ContentPage, 0, len(result.Pages)),
		NextPageToken: result.NextPageToken,
		TotalSize:     result.TotalSize,
	}
	for _, item := range result.Pages {
		response.Pages = append(response.Pages, openapi.ContentPage{
			Id:          &item.ID,
			Slug:        item.Slug,
			Title:       item.Title,
			Html:        &item.HTML,
			PublishedAt: item.PublishedAt,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	return response, nil
}
