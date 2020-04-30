//go:generate gex mockgen -source=clock.go -package usecases -destination=clock_mock.go

package usecases

import (
	"time"
)

// Clock abstracts away getting the current time, so it can be mocked in tests
type Clock interface {
	Now() time.Time
}
