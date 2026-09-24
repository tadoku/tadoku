package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/asyncwork"
	"github.com/tadoku/tadoku/services/tadoku-api/storage/postgres/asyncoutbox"
)

func TestDispatchRejectsInvalidKnownPayload(t *testing.T) {
	app := NewApplication(nil)
	for _, task := range []asyncoutbox.ClaimedTask{
		{Type: asyncwork.InvalidateContest, Payload: []byte(`{"contest_id":"bad"}`)},
		{Type: asyncwork.InvalidateOfficial, Payload: []byte(`{"year":0}`)},
	} {
		err := app.Dispatch(context.Background(), task)
		var permanent *PermanentError
		if !errors.As(err, &permanent) {
			t.Errorf("%s: got %v; want permanent error", task.Type, err)
		}
	}
}

func TestDispatchLeavesUnknownTypeUnhandled(t *testing.T) {
	err := NewApplication(nil).Dispatch(context.Background(), asyncoutbox.ClaimedTask{Type: "future.task.v1"})
	var unknown *UnknownTypeError
	if !errors.As(err, &unknown) {
		t.Errorf("got %v; want unknown type error", err)
	}
}
