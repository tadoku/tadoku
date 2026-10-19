package main

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestDevelopmentAccountsCommand(t *testing.T) {
	db := accessTokenDatabase(t)
	admin, reader := uuid.NewString(), uuid.NewString()
	args := []string{"development-accounts", "--tenant", "e2e/tokens-0123abcd", "--admin-id", admin, "--reader-id", reader}
	t.Setenv("TADOKU_DEV_ADMIN_PASSWORD", "tadoku")
	t.Setenv("TADOKU_DEV_READER_PASSWORD", "tadoku")

	if err := run(t.Context(), args); err != nil {
		t.Fatalf("seed native development accounts: %v", err)
	}
	snapshot := func() []string {
		t.Helper()
		rows, err := db.Pool.Query(t.Context(), `select row_to_json(a)::text from accounts a order by tenant,user_id`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				t.Fatal(err)
			}
			out = append(out, s)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return out
	}
	before := snapshot()
	for range 2 {
		if err := run(t.Context(), args); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(before, snapshot()) {
		t.Fatal("rerun changed existing accounts")
	}

	for _, fixture := range []struct{ id, email, name string }{
		{admin, "dev@tadoku.app", "Dev Admin"},
		{reader, "reader@tadoku.app", "Dev Reader"},
	} {
		var email, name, hash string
		err := db.Pool.QueryRow(t.Context(), `select a.email,u.display_name,a.password_hash
   from accounts a join users u on (u.tenant,u.id)=(a.tenant,a.user_id)
   where a.tenant='e2e/tokens-0123abcd' and a.user_id=$1`, fixture.id).Scan(&email, &name, &hash)
		if err != nil || email != fixture.email || name != fixture.name ||
			bcrypt.CompareHashAndPassword([]byte(hash), []byte("tadoku")) != nil {
			t.Fatalf("native fixture %s has wrong identity or password (error=%v)", fixture.name, err)
		}
	}
	t.Setenv("TADOKU_DEV_READER_PASSWORD", "different-fixture-password")
	if err := run(t.Context(), args); err == nil {
		t.Fatal("overwrote an existing fixture password")
	}
	if !reflect.DeepEqual(before, snapshot()) {
		t.Fatal("conflict changed accounts")
	}

	for _, blocked := range []string{"tadoku/prod", "e2e/missing-0123abcd"} {
		denied := append([]string{}, args...)
		denied[2] = blocked
		if err := run(t.Context(), denied); err == nil {
			t.Fatalf("seed accepted %s", blocked)
		}
	}
	t.Setenv("TENANT_PRODUCTION", "true")
	if err := run(t.Context(), args); err == nil {
		t.Fatal("production accepted development fixtures")
	}
	if !reflect.DeepEqual(before, snapshot()) {
		t.Fatal("denied invocation changed accounts")
	}
	var outside int
	err := db.Pool.QueryRow(
		t.Context(),
		`select count(*) from accounts where tenant<>'e2e/tokens-0123abcd'`,
	).Scan(&outside)
	if err != nil || outside != 0 {
		t.Fatalf("seed changed another tenant: count=%d error=%v", outside, err)
	}
}
