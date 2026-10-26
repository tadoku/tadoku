package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestDeleteAccountCommandFilesRequest(t *testing.T) {
	db := workerTestDatabase(t)
	const userID = "11111111-1111-4111-8111-111111111111"
	const operatorID = "22222222-2222-4222-8222-222222222222"
	_, err := db.Pool.Exec(t.Context(), `insert into users (tenant, id, display_name)
		values ('tadoku/prod', $1, 'Reader One')`, userID)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, nil))
	args := []string{"--user", userID, "--actor", operatorID, "--reason", " owner asked by email "}

	if err := deleteAccount(t.Context(), args, logger); err != nil {
		t.Fatal(err)
	}

	var jobID int64
	var locked bool
	var actor, source, reason string
	err = db.Pool.QueryRow(t.Context(), `select jobs.id, users.deletion_locked_at is not null,
			audit.user_id::text, audit.metadata->>'source', audit.description
		from users
		join jobs on jobs.tenant = users.tenant
			and jobs.task_type = 'account.delete.v1'
			and jobs.payload->>'user_id' = users.id::text
		join moderation_audit_log as audit on audit.tenant = users.tenant
			and audit.action = 'request_account_deletion'
			and audit.metadata->>'target_user_id' = users.id::text
		where users.tenant = 'tadoku/prod' and users.id = $1`, userID).
		Scan(&jobID, &locked, &actor, &source, &reason)
	if err != nil {
		t.Fatalf("command did not lock, audit and enqueue the deletion: %v", err)
	}
	if !locked || actor != operatorID || source != "operator" || reason != "owner asked by email" {
		t.Errorf("filing: locked=%t actor=%q source=%q reason=%q", locked, actor, source, reason)
	}
	if !strings.Contains(output.String(), fmt.Sprintf("job_id=%d", jobID)) {
		t.Errorf("command output %q does not name job %d", output.String(), jobID)
	}

	if err := deleteAccount(t.Context(), args, logger); err == nil {
		t.Error("command filed a second deletion for the same user")
	}
	for _, invalid := range [][]string{
		{"--user", userID, "--actor", operatorID},
		{"--user", "not-a-uuid", "--actor", operatorID, "--reason", "owner asked"},
		{"--user", userID, "--reason", "owner asked"},
		{"--user", userID, "--actor", operatorID, "--reason", "owner asked", "extra"},
	} {
		if err := deleteAccount(t.Context(), invalid, logger); err == nil {
			t.Errorf("command accepted %q", invalid)
		}
	}
}
