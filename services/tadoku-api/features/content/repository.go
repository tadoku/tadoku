package content

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) ListActiveAnnouncements(ctx context.Context, namespace string, cutoff time.Time, limit int32) ([]Announcement, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).ListActiveAnnouncements(ctx, queries.ListActiveAnnouncementsParams{
		Namespace: namespace, Cutoff: pgtype.Timestamp{Time: cutoff, Valid: true}, ResultLimit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list active announcements: %w", err)
	}
	result := make([]Announcement, 0, len(rows))
	for _, row := range rows {
		var href *string
		if row.Href.Valid {
			href = &row.Href.String
		}
		result = append(result, Announcement{
			ID: uuid.UUID(row.ID.Bytes), Namespace: row.Namespace, Title: row.Title,
			Content: row.Content, Style: row.Style, Href: href,
			StartsAt: row.StartsAt.Time, EndsAt: row.EndsAt.Time,
			CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
		})
	}
	return result, nil
}
