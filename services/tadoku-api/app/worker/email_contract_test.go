package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
)

func TestEmailSendEncoding(t *testing.T) {
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	payload := jobs.SendEmailV1{EmailID: id}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"email_id":"11111111-1111-4111-8111-111111111111"}` {
		t.Fatalf("payload contains more than email_id: %s", raw)
	}
	var received uuid.UUID
	registration := handle(func(_ context.Context, job jobs.SendEmailV1) error {
		received = job.EmailID
		return nil
	}, policy{Concurrency: 1, Timeout: time.Second, MaxAttempts: 1})
	if registration.err != nil {
		t.Fatal(registration.err)
	}
	if registration.spec.typeName != jobs.Type("email.send.v1") {
		t.Fatalf("name=%s", registration.spec.typeName)
	}

	if err := registration.invoke(t.Context(), raw); err != nil {
		t.Fatal(err)
	}
	if received != id {
		t.Fatalf("decoded id=%s", received)
	}
	for _, invalid := range []string{
		`{}`,
		`{"email_id":null}`,
		`{"email_id":"bad"}`,
		`{"email_id":"11111111-1111-4111-8111-111111111111","recipient":"secret@example.test"}`,
	} {
		received = uuid.Nil
		if err := registration.invoke(t.Context(), []byte(invalid)); err == nil || received != uuid.Nil {
			t.Errorf("invalid payload reached handler: %s error=%v", invalid, err)
		}
	}
}
