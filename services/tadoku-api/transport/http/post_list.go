package http

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func (s *server) ContentPostList(
	ctx context.Context,
	request openapi.ContentPostListRequestObject,
) (openapi.ContentPostListResponseObject, error) {
	pageSize, page := 0, 0
	if request.Params.PageSize != nil {
		pageSize = *request.Params.PageSize
	}
	if request.Params.Page != nil {
		page = *request.Params.Page
	}

	result, err := s.application.ListPosts(ctx, request.Namespace, request.Params.IncludeDrafts, pageSize, page)
	if err != nil {
		s.logOperationError(ctx, "list posts", err)
		return nil, err
	}

	response := openapi.ContentPostList200JSONResponse{
		Posts:         make([]openapi.ContentPost, 0, len(result.Posts)),
		NextPageToken: result.NextPageToken,
		TotalSize:     result.TotalSize,
	}
	for _, item := range result.Posts {
		response.Posts = append(response.Posts, openapi.ContentPost{
			Id:          &item.ID,
			Slug:        item.Slug,
			Title:       item.Title,
			Content:     item.Content,
			PublishedAt: item.PublishedAt,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	return response, nil
}
