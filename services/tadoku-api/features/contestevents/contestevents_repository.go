package contestevents

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/contestevents"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) Insert(ctx context.Context, events []newEvent, occurredAt time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	if _, active := executor.(pgx.Tx); !active {
		return errors.New("contest event publication requires a transaction")
	}

	params := queries.InsertContestEventsParams{OccurredAt: postgres.Timestamptz(occurredAt)}
	for _, event := range events {
		params.Ids = append(params.Ids, postgres.UUID(event.id))
		params.ContestIds = append(params.ContestIds, postgres.UUID(event.change.ContestID))
		params.LogIds = append(params.LogIds, postgres.UUID(event.change.LogID))
		params.Events = append(params.Events, string(event.change.Type))
	}
	return queries.New(executor).InsertContestEvents(ctx, params)
}

func (r *Repository) Read(ctx context.Context, parameters readParameters) (*Page, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}
	rows, err := queries.New(executor).ReadContestEvents(ctx, queries.ReadContestEventsParams{
		ContestID: postgres.UUID(parameters.contestID),
		CursorID:  postgres.NullableUUID(parameters.cursorID),
		FromNow:   parameters.fromNow,
		PageSize:  parameters.limit,
	})
	if err != nil {
		return nil, err
	}

	page := &Page{Events: make([]Event, 0, len(rows)), NextCursor: "0"}
	for _, row := range rows {
		if !row.CursorValid {
			return nil, errx.NewInvalidInputError("cursor does not belong to this contest")
		}
		if row.CheckpointID.Valid {
			page.NextCursor = uuid.UUID(row.CheckpointID.Bytes).String()
		}
		if row.ID.Valid {
			page.Events = append(page.Events, Event{
				ID:         uuid.UUID(row.ID.Bytes),
				LogID:      uuid.UUID(row.LogID.Bytes),
				Type:       contestevent.Type(row.Event.String),
				OccurredAt: row.OccurredAt.Time,
			})
		}
	}
	return page, nil
}
