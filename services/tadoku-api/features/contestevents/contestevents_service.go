package contestevents

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// Append requires the source mutation's transaction and operation timestamp.
func (s *Service) Append(ctx context.Context, changes []contestevent.Change, occurredAt time.Time) error {
	for _, change := range changes {
		if err := s.repository.Insert(ctx, uuid.New(), change, occurredAt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context, contestID uuid.UUID, cursor string, limit int) (*Page, error) {
	if limit < 1 || limit > 100 {
		return nil, errx.NewInvalidInputError("limit must be between 1 and 100")
	}
	if cursor == "" || len(cursor) > 36 {
		return nil, errx.NewInvalidInputError("cursor must be 0, now, or an event UUID")
	}

	parameters := readParameters{contestID: contestID, limit: int32(limit + 1)}
	switch cursor {
	case "0":
	case "now":
		parameters.fromNow = true
	default:
		id, err := uuid.Parse(cursor)
		if err != nil || len(cursor) != 36 {
			return nil, errx.NewInvalidInputError("cursor must be 0, now, or an event UUID")
		}
		parameters.cursorID = &id
	}

	page, err := s.repository.Read(ctx, parameters)
	if err != nil {
		return nil, err
	}
	if len(page.Events) > limit {
		page.Events = page.Events[:limit]
		page.HasMore = true
	}
	if len(page.Events) > 0 {
		page.NextCursor = page.Events[len(page.Events)-1].ID.String()
	}
	return page, nil
}
