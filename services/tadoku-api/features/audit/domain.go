package audit

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ActorID     uuid.UUID
	Action      string
	Metadata    map[string]any
	Description string
	recordedAt  time.Time
}
