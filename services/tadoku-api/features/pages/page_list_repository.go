package pages

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PagesRepository) ListPages(ctx context.Context, namespace string, includeDrafts bool, limit int32, offset int64) ([]Page, int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	rows, err := queries.New(executor).ListPages(ctx, queries.ListPagesParams{
		Namespace:     namespace,
		IncludeDrafts: includeDrafts,
		ResultLimit:   limit,
		StartFrom:     offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list pages: %w", err)
	}

	result := make([]Page, 0, len(rows))
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
		result = append(result, Page{
			ID:          uuid.UUID(row.ID.Bytes),
			Namespace:   row.Namespace.String,
			Slug:        row.Slug.String,
			Title:       row.Title.String,
			HTML:        row.Html.String,
			PublishedAt: publishedAt,
			CreatedAt:   &row.CreatedAt.Time,
			UpdatedAt:   &row.UpdatedAt.Time,
		})
	}

	return result, total, nil
}
