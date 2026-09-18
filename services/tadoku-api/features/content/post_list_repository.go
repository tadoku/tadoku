package content

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PostsRepository) ListPosts(ctx context.Context, namespace string, includeDrafts bool, cutoff time.Time, limit int32, offset int64) ([]Post, int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	rows, err := queries.New(executor).ListPosts(ctx, queries.ListPostsParams{
		Namespace:     namespace,
		IncludeDrafts: includeDrafts,
		Cutoff:        postgres.Timestamp(cutoff),
		ResultLimit:   limit,
		StartFrom:     offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}

	result := make([]Post, 0, len(rows))
	var total int
	for _, row := range rows {
		total = int(row.TotalSize)
		if !row.ID.Valid {
			continue
		}

		var publishedAt *time.Time
		if row.PublishedAt.Valid {
			publishedAt = &row.PublishedAt.Time
		}

		result = append(result, Post{
			ID:          uuid.UUID(row.ID.Bytes),
			Namespace:   row.Namespace.String,
			Slug:        row.Slug.String,
			Title:       row.Title.String,
			Content:     row.Content.String,
			PublishedAt: publishedAt,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}

	return result, total, nil
}
