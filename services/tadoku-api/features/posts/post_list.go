package posts

import (
	"context"
	"math"
	"strconv"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type PostList struct {
	Posts         []Post
	TotalSize     int
	NextPageToken string
}

func (s *Service) ListPosts(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*PostList, error) {
	if namespace == "" {
		return nil, ErrInvalidNamespace
	}
	if pageSize < 0 || page < 0 {
		return nil, ErrInvalidPagination
	}
	if pageSize == 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := int64(page)
	if page > math.MaxInt64/pageSize {
		offset = math.MaxInt64
	} else {
		offset *= int64(pageSize)
	}

	posts, totalSize, err := s.posts.ListPosts(ctx, namespace, includeDrafts, timex.Now(), int32(pageSize), offset)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if int64(pageSize) < int64(totalSize)-offset {
		nextPageToken = strconv.Itoa(page + 1)
	}

	return &PostList{
		Posts:         posts,
		TotalSize:     totalSize,
		NextPageToken: nextPageToken,
	}, nil
}
