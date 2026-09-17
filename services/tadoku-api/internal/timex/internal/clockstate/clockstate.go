package clockstate

import (
	"sync/atomic"
	"time"
)

var fixedTime atomic.Pointer[time.Time]

func Fixed() (time.Time, bool) {
	fixed := fixedTime.Load()
	if fixed == nil {
		return time.Time{}, false
	}
	return *fixed, true
}

func Set(instant time.Time) bool {
	return fixedTime.CompareAndSwap(nil, &instant)
}

func Clear() {
	fixedTime.Store(nil)
}
