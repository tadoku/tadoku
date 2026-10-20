package main_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
)

func TestImportOperationRequiresPositiveApplyCount(t *testing.T) {
	for _, expect := range []int{0, -1} {
		t.Run(fmt.Sprintf("expect-%d", expect), func(t *testing.T) {
			f := newFixture(t)
			before := f.snapshot(t)
			var output bytes.Buffer

			err := identity.ImportKratosAccounts(t.Context(), f.source.AppPool, f.target.AppPool, expect, true, &output)
			if err == nil {
				t.Errorf("operation accepted apply with expect=%d: %s", expect, output.String())
			} else if !strings.Contains(err.Error(), "positive") || !strings.Contains(err.Error(), "count") {
				t.Errorf("operation rejected apply for an unrelated reason: %v", err)
			}
			f.unchanged(t, before)
		})
	}
}

func TestImportOperationDryRunAndApply(t *testing.T) {
	f := newFixture(t)
	before := f.snapshot(t)
	var output bytes.Buffer

	err := identity.ImportKratosAccounts(t.Context(), f.source.AppPool, f.target.AppPool, 0, false, &output)
	if err != nil || !strings.Contains(output.String(), "dry_run=true insert_users=2 insert_accounts=2") {
		t.Fatalf("operation dry run failed: error=%v output=%s", err, output.String())
	}
	f.unchanged(t, before)

	output.Reset()
	err = identity.ImportKratosAccounts(t.Context(), f.source.AppPool, f.target.AppPool, 2, true, &output)
	if err != nil || !strings.Contains(output.String(), "committed_accounts=2 target_accounts=2") {
		t.Fatalf("operation apply failed: error=%v output=%s", err, output.String())
	}
	var count int
	err = f.target.Pool.QueryRow(t.Context(), `select count(*)
		from accounts
		where tenant = 'tadoku/prod'`).Scan(&count)
	if err != nil || count != 2 {
		t.Fatal("operation did not persist both accounts")
	}
}

func TestImportDeletedLegacyUsersWithoutAccounts(t *testing.T) {
	for _, column := range []string{"deleted_at", "deletion_locked_at"} {
		t.Run(column, func(t *testing.T) {
			f := newFixture(t)
			f.sql(t, false, `insert into users (tenant, id, display_name, created_at)
				values ('tadoku/prod', $1, 'Stale legacy name', '2022-06-15')`, secondID)
			f.sql(t, false, "update users set "+column+" = now() where tenant = 'tadoku/prod' and id = $1", secondID)
			var count int
			err := f.target.Pool.QueryRow(t.Context(), "select count(*) from accounts").Scan(&count)
			if err != nil || count != 0 {
				t.Fatal("legacy deletion fixture must have no existing accounts")
			}
			before := f.snapshot(t)

			for _, args := range [][]string{{"--dry-run"}, {"--apply", "--expect-count", "2"}} {
				out, err := f.command(t, args...)
				if err == nil || !strings.Contains(out, "target conflict user_id="+secondID+" field=deletion_state") {
					t.Errorf("deleted legacy user accepted or failed for another reason: error=%v output=%s", err, out)
				}
				f.unchanged(t, before)
			}
		})
	}
}
