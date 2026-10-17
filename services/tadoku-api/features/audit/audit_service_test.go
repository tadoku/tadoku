package audit

import (
	"context"
	"testing"
)

func TestRecordContextOutlivesTheCaller(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	recording, release := recordContext(ctx)
	defer release()

	if err := recording.Err(); err != nil {
		t.Errorf("recording context error = %v, want nil", err)
	}
	if _, ok := recording.Deadline(); !ok {
		t.Error("recording context has no deadline")
	}
}
