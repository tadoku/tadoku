package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/query"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) FetchUserTags(ctx context.Context, req *query.FetchUserTagsRequest) (*query.FetchUserTagsResponse, error) {
	// Set defaults
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	userTags, err := r.q.ListUserTags(ctx, postgres.ListUserTagsParams{
		UserID:  req.UserID,
		Column2: req.Prefix,
		Limit:   int32(limit),
		Offset:  int32(req.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("could not fetch user tags: %w", err)
	}

	defaultTags, err := r.q.ListDefaultTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not fetch default tags: %w", err)
	}

	tags := make([]query.UserTag, len(userTags))
	for i, t := range userTags {
		tags[i] = query.UserTag{
			Tag:        t.Tag,
			UsageCount: t.UsageCount,
		}
	}

	return &query.FetchUserTagsResponse{
		Tags:        tags,
		DefaultTags: defaultTags,
	}, nil
}
