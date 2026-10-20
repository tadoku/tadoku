package main_test

import (
	"strings"
	"testing"
)

const lateControlID = "ffffffff-ffff-4fff-8fff-ffffffffffff"
const ideographicControlID = "33333333-3333-4333-8333-333333333333"

// Each address normalizes to a distinct valid login key, so the importer admits it, but recovery and
// email-change notices address the stored value and the email writer rejects it as a recipient.
var incompatibleRecipients = []struct{ id, email string }{
	{"aaaaaaaa-aaaa-4aaa-8aaa-000000000001", "\vvertical@example.test"},
	{"aaaaaaaa-aaaa-4aaa-8aaa-000000000002", " space@example.test"},
	{"aaaaaaaa-aaaa-4aaa-8aaa-000000000003", "tab@example.test\t"},
	{"aaaaaaaa-aaaa-4aaa-8aaa-000000000004", "crlf@example.test\r\n"},
	{"aaaaaaaa-aaaa-4aaa-8aaa-000000000005", strings.Repeat("a", 316) + "K@x"},
	{"aaaaaaaa-aaaa-4aaa-8aaa-000000000006", strings.Repeat("b", 317) + "İ@x"},
	{"aaaaaaaa-aaaa-4aaa-8aaa-000000000007", strings.Repeat("　", 350) + "wide@example.test"},
}

func TestImportRefusesIncompatibleRecipientsBeforeTargetWrites(t *testing.T) {
	f := newFixture(t)
	f.addIdentity(t, ideographicControlID, "　ideographic@example.test", "Fixture reader", "active")
	for _, record := range incompatibleRecipients {
		f.addIdentity(t, record.id, record.email, "Fixture reader", "active")
	}
	f.addIdentity(t, lateControlID, "Late.Control@Example.Test", "Fixture reader", "active")
	writes := installTargetWriteProbe(t, f)
	before := f.snapshot(t)

	var want []string
	for _, record := range incompatibleRecipients {
		want = append(want, "preflight user_id="+record.id+" field=email reason=recipient_rule")
	}
	for _, args := range [][]string{{"--dry-run"}, {"--apply", "--expect-count", "11"}} {
		out, err := f.command(t, args...)
		if err == nil {
			t.Errorf("%v admitted a snapshot with incompatible recipients", args)
		}
		if got := diagnostics(out); strings.Join(got, "\n") != strings.Join(want, "\n") {
			t.Errorf("%v reported %d diagnostics %q, want every incompatible UUID in order", args, len(got), got)
		}
		if strings.Contains(out, "example.test") || strings.Contains(out, "@x") || strings.Contains(out, "Fixture") {
			t.Fatal("CLI output exposed a source email or display name (output suppressed)")
		}
		if count := writes(); count != 0 {
			t.Errorf("%v attempted %d target row writes before refusing", args, count)
		}
		f.unchanged(t, before)
	}
	if t.Failed() {
		t.FailNow()
	}

	f.sql(t, true, "delete from identities where id::text like 'aaaaaaaa-%'")
	out, err := f.command(t, "--apply", "--expect-count", "4")
	if err != nil || !strings.Contains(out, "committed_accounts=4") {
		t.Fatalf("compatible snapshot was refused: error=%v output=%s", err, out)
	}
	if writes() == 0 {
		t.Fatal("target write probe did not observe the compatible import")
	}

	var preserved bool
	err = f.target.Pool.QueryRow(t.Context(), `select count(*) = 2 and bool_and(email = case user_id
			when $1::uuid then $3
			else 'Late.Control@Example.Test'
		end
		and email_normalized = case user_id
			when $1::uuid then 'ideographic@example.test'
			else 'late.control@example.test'
		end)
		from accounts
		where tenant = 'tadoku/prod' and user_id in ($1, $2)`,
		ideographicControlID,
		lateControlID,
		"　ideographic@example.test",
	).Scan(&preserved)
	if err != nil || !preserved {
		t.Fatal("compatible controls lost their stored email or normalized lookup key")
	}
}

func diagnostics(output string) []string {
	var result []string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "preflight user_id=") {
			result = append(result, line)
		}
	}
	return result
}

// Test safety: the probe counts attempted row writes on the disposable target with a sequence, which a rollback
// does not rewind, so a refusal that wrote and rolled back still fails.
func installTargetWriteProbe(t *testing.T, f *fixture) func() int64 {
	t.Helper()

	f.sql(t, false, `create sequence fixture_target_writes;
		create function fixture_count_target_write() returns trigger
		language plpgsql
		as $$ begin perform nextval('fixture_target_writes'); return null; end $$;
		create trigger fixture_users_writes after insert or update or delete on users
		for each row execute function fixture_count_target_write();
		create trigger fixture_accounts_writes after insert or update or delete on accounts
		for each row execute function fixture_count_target_write()`)
	f.sql(t, false, `grant usage on sequence fixture_target_writes to "`+f.target.AppPool.Config().ConnConfig.User+`"`)
	return func() int64 {
		t.Helper()

		var calls int64
		err := f.target.Pool.QueryRow(t.Context(), `select case when is_called then last_value else 0 end
			from fixture_target_writes`).Scan(&calls)
		if err != nil {
			t.Fatal("read target write probe failed")
		}
		return calls
	}
}
