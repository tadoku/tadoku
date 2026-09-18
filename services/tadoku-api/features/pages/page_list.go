package pages

import (
	"context"
	"math"
	"strconv"
)

type PageList struct {
	Pages         []Page
	TotalSize     int
	NextPageToken string
}

func (s *Service) ListPages(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*PageList, error) {
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

	pages, totalSize, err := s.pages.ListPages(ctx, namespace, includeDrafts, int32(pageSize), offset)
	if err != nil {
		return nil, err
	}

	nextPageToken := ""
	if int64(pageSize) < int64(totalSize)-offset {
		nextPageToken = strconv.Itoa(page + 1)
	}

	return &PageList{Pages: pages, TotalSize: totalSize, NextPageToken: nextPageToken}, nil
}
