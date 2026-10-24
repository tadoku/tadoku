package contestevents

import (
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/contestevent"
)

type Event struct {
	ID         uuid.UUID
	LogID      uuid.UUID
	Type       contestevent.Type
	OccurredAt time.Time
}

type Page struct {
	Events     []Event
	NextCursor string
	HasMore    bool
}

type readParameters struct {
	contestID uuid.UUID
	cursorID  *uuid.UUID
	fromNow   bool
	limit     int32
}
