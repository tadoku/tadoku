package jobqueue

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
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

func validateErrorCode(code string) error {
	if len(code) == 0 || len(code) > 100 {
		return errx.NewInvalidInputError("job error code must be 1 to 100 characters")
	}
	for _, r := range code {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return errx.NewInvalidInputError("job error code must contain only lowercase letters, digits, or underscores")
		}
	}
	return nil
}
