package usecases

import (
	"time"
)

// Clock abstracts away getting the current time, so it can be mocked in tests
type Clock interface {
	Now() time.Time
}
