package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/mail"
	"github.com/tadoku/tadoku/services/tadoku-api/features/email"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func recoveryMail() mail.Message {
	return mail.Message{Template: mail.AccountRecovery, Recipient: "reader@example.test", Params: map[string]string{
		"recovery_url": "https://account.tadoku.app/identity/recovery/secret",
	}}
}

func emailService() *email.Service {
	return email.NewPersistentService(email.NewRepository(api.db.AppPool), nil)
}

func writeEmail(t *testing.T, ctx context.Context, service *email.Service, message mail.Message) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := postgres.RunInTransaction(ctx, api.db.AppPool, func(ctx context.Context) error {
		result, err := service.Write(ctx, message)
		if err != nil {
			return err
		}
		id = result.EmailID
		return jobqueue.NewService(jobqueue.NewRepository(api.db.AppPool)).Enqueue(ctx, result.Jobs...)
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestEmailWriteValidation(t *testing.T) {
	cases := []struct {
		name    string
		message mail.Message
	}{
		{"unknown", mail.Message{Template: "unknown", Recipient: "reader@example.test"}},
		{"missing", mail.Message{Template: mail.AccountRecovery, Recipient: "reader@example.test"}},
		{"unexpected", mail.Message{Template: mail.AccountRecovery, Recipient: "reader@example.test", Params: map[string]string{
			"recovery_url": "https://tadoku.app/recover", "extra": "secret"}}},
		{"empty", mail.Message{Template: mail.AccountRecovery, Recipient: "reader@example.test", Params: map[string]string{"recovery_url": ""}}},
		{"email changed missing", mail.Message{Template: mail.AccountEmailChanged, Recipient: "reader@example.test", Params: map[string]string{"new_email": "new@example.test"}}},
	}
	for _, raw := range []string{"https://evil.test/recover", "javascript:alert(1)", "https://tadoku.app.evil.test/recover", "https://tadoku.app@evil.test/", "http://account.tadoku.app/recover", "https://evil.preview.tadoku.app.evil.test/"} {
		m := recoveryMail()
		m.Params["recovery_url"] = raw
		cases = append(cases, struct {
			name    string
			message mail.Message
		}{raw, m})
	}
	for _, recipient := range []string{"bad", "a@example.test\r\nBcc: b@example.test", "Name <a@example.test>"} {
		m := recoveryMail()
		m.Recipient = recipient
		cases = append(cases, struct {
			name    string
			message mail.Message
		}{recipient, m})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tenant.WithKey(t.Context(), tenant.Production())
			err := postgres.RunInTransaction(ctx, api.db.AppPool, func(ctx context.Context) error {
				_, err := emailService().Write(ctx, tc.message)
				return err
			})
			if errx.KindOf(err) != errx.InvalidInput {
				t.Errorf("write error=%v kind=%v, want invalid input", err, errx.KindOf(err))
			}
		})
	}
}

func TestEmailWriteTransactionAndPayload(t *testing.T) {
	api.reset(t, "")
	ctx := tenant.WithKey(t.Context(), tenant.Production())
	service := emailService()
	if _, err := service.Write(ctx, recoveryMail()); err == nil {
		t.Error("write without business transaction succeeded")
	}
	aborted := errors.New("business operation failed")
	err := postgres.RunInTransaction(ctx, api.db.AppPool, func(ctx context.Context) error {
		result, err := service.Write(ctx, recoveryMail())
		if err != nil {
			return err
		}
		if err = jobqueue.NewService(jobqueue.NewRepository(api.db.AppPool)).Enqueue(ctx, result.Jobs...); err != nil {
			return err
		}
		return aborted
	})
	if !errors.Is(err, aborted) {
		t.Fatalf("rollback cause=%v", err)
	}
	var count int
	if err := api.db.Pool.QueryRow(ctx, "select (select count(*) from emails)+(select count(*) from jobs)").Scan(&count); err != nil || count != 0 {
		t.Fatalf("rollback count=%d err=%v", count, err)
	}
	id := writeEmail(t, ctx, service, recoveryMail())
	var payload []byte
	if err := api.db.Pool.QueryRow(ctx, "select payload from jobs where task_type='email.send.v1'").Scan(&payload); err != nil {
		t.Fatal(err)
	}
	var fields map[string]string
	if err := json.Unmarshal(payload, &fields); err != nil || len(fields) != 1 || fields["email_id"] != id.String() {
		t.Fatalf("payload=%s err=%v", payload, err)
	}
	row, err := service.Find(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if row.Recipient != "reader@example.test" || !strings.Contains(row.Body, recoveryMail().Params["recovery_url"]) || row.Subject == "" {
		t.Errorf("rendered row=%+v", row)
	}
	if _, ok := row.Delivery.(email.Pending); !ok {
		t.Errorf("delivery=%T", row.Delivery)
	}
	changed := mail.Message{Template: mail.AccountEmailChanged, Recipient: "old@example.test", Params: map[string]string{"new_email": "new@example.test", "changed_at": "2026-10-19T12:00:00Z"}}
	changedID := writeEmail(t, ctx, service, changed)
	row, err = service.Find(ctx, changedID)
	if err != nil || !strings.Contains(row.Body, "new@example.test") || !strings.Contains(row.Body, changed.Params["changed_at"]) {
		t.Fatalf("changed row=%+v err=%v", row, err)
	}
}
