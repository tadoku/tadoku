package jobqueue

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
)

var (
	ErrNotFailed = errors.New("job is not failed or supported")
)

type ClaimedJob struct {
	ID             int64
	Type           jobs.Type
	Payload        json.RawMessage
	Attempts       int
	Token          uuid.UUID
	LeaseExpiresAt time.Time
	Reclaimed      bool
}

type Stats struct {
	Pending     int64
	Failed      int64
	OldestDueAt *time.Time
}

type UnsupportedStats struct {
	Pending     int64
	Running     int64
	Failed      int64
	OldestDueAt *time.Time
}
