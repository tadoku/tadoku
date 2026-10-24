package contestevent

import "github.com/google/uuid"

type Type string

const (
	Submitted Type = "log.submitted"
	Updated   Type = "log.updated"
	Withdrawn Type = "log.withdrawn"
)

type Change struct {
	ContestID uuid.UUID
	LogID     uuid.UUID
	Type      Type
}
