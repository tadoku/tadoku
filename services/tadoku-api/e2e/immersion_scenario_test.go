package e2e_test

import (
	"context"
	"testing"
	"time"
)

// resetAfter resets the database when the serial top-level test t ends. Call it from a serial test that writes shared
// reference data, such as languages, so the parallel scenarios that start after every serial test read the baseline.
func resetAfter(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := api.db.Reset(ctx); err != nil {
			t.Error(err)
		}
	})
}
