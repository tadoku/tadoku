package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func accessTokenDatabase(t *testing.T) *testpostgres.Database {
	t.Helper()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	u, err := url.Parse(db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	for key, value := range map[string]string{
		"HOST": u.Hostname(), "PORT": u.Port(), "DATABASE": strings.TrimPrefix(u.Path, "/"),
		"USER": u.User.Username(), "PASSWORD": password, "SSLMODE": "disable",
	} {
		t.Setenv("TENANT_POSTGRES_"+key, value)
	}
	t.Setenv("TENANT_PRODUCTION", "false")
	_, err = db.Pool.Exec(t.Context(), `insert into tenants(key,kind)
 values ('e2e/tokens-0123abcd','test'),
 ('e2e/other-0123abcd','test')`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func tokenCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = previous; read.Close() }()
	err = run(t.Context(), append([]string{"access-token"}, args...))
	write.Close()
	raw, readErr := io.ReadAll(read)
	if readErr != nil {
		t.Fatal(readErr)
	}
	return strings.TrimSpace(string(raw)), err
}

func TestAccessTokenCommandLifecycle(t *testing.T) {
	db := accessTokenDatabase(t)
	raw, err := tokenCommand(t,
		"create", "--tenant", "e2e/tokens-0123abcd", "--scope", "email:read",
		"--name", "reader", "--expires-in", "720h",
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !strings.HasPrefix(raw, "tdk_at_e2e_tokens-0123abcd_") || len(strings.Fields(raw)) != 1 {
		t.Fatal("create did not print one tenant-prefixed token")
	}
	hash := sha256.Sum256([]byte(raw))
	var stored string
	var id uuid.UUID
	err = db.Pool.QueryRow(t.Context(),
		"select id, token_hash from access_tokens where tenant='e2e/tokens-0123abcd'",
	).Scan(&id, &stored)
	if err != nil {
		t.Fatal(err)
	}
	if stored != hex.EncodeToString(hash[:]) || stored == raw {
		t.Fatal("database did not store only the token SHA-256 hash")
	}
	list, err := tokenCommand(t, "list", "--tenant", "e2e/tokens-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	var metadata []struct {
		ID        uuid.UUID
		Name      string
		ExpiresAt time.Time
		RevokedAt *time.Time
	}
	if json.Unmarshal([]byte(list), &metadata) != nil || len(metadata) != 1 ||
		metadata[0].ID != id || metadata[0].Name != "reader" {
		t.Fatal("list did not return token metadata")
	}
	if strings.Contains(list, raw) || strings.Contains(list, stored) {
		t.Fatal("list revealed credentials")
	}
	if _, err := tokenCommand(t, "revoke", "--tenant", "e2e/other-0123abcd", "--id", id.String()); err == nil {
		t.Error("other tenant revoked token")
	}
	for range 2 {
		if _, err := tokenCommand(t, "revoke", "--tenant", "e2e/tokens-0123abcd", "--id", id.String()); err != nil {
			t.Fatal(err)
		}
	}
	list, err = tokenCommand(t, "list", "--tenant", "e2e/tokens-0123abcd")
	if err != nil || json.Unmarshal([]byte(list), &metadata) != nil || metadata[0].RevokedAt == nil {
		t.Fatal("revocation not persisted in list")
	}
}

func TestAccessTokenCommandLimits(t *testing.T) {
	db := accessTokenDatabase(t)
	create := []string{
		"create", "--tenant", "e2e/tokens-0123abcd", "--scope", "email:read",
		"--name", "reader", "--expires-in", "1h",
	}
	if _, err := db.Pool.Exec(t.Context(), `insert into access_tokens(tenant,name,token_hash,scopes,expires_at)
  select 'e2e/tokens-0123abcd','existing',n::text,array['email:read'],now()+interval '1 day'
  from generate_series(1,10) n`); err != nil {
		t.Fatal(err)
	}
	if _, err := tokenCommand(t, create...); err == nil || !strings.Contains(err.Error(), "10 active") {
		t.Errorf("eleventh token: %v, want active-token limit", err)
	}
	create[len(create)-1] = "720h1ns"
	if _, err := tokenCommand(t, create...); err == nil || !strings.Contains(err.Error(), "30 days") {
		t.Errorf("overlong duration: %v, want duration limit", err)
	}
	t.Setenv("TENANT_PRODUCTION", "true")
	create[2] = "tadoku/prod"
	create[len(create)-1] = "1h"
	if _, err := tokenCommand(t, create...); err == nil || !strings.Contains(err.Error(), "production") {
		t.Errorf("production token: %v, want production refusal", err)
	}
}

func TestAccessTokenCommandConcurrentLimit(t *testing.T) {
	db := accessTokenDatabase(t)
	if _, err := db.Pool.Exec(t.Context(), `insert into access_tokens(tenant,name,token_hash,scopes,expires_at)
  select 'e2e/tokens-0123abcd','existing',n::text,array['email:read'],now()+interval '1 day'
  from generate_series(1,9) n`); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = write
	defer func() { os.Stdout = previous; read.Close() }()
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for range 8 {
		wg.Go(func() {
			results <- run(context.Background(), []string{
				"access-token", "create", "--tenant", "e2e/tokens-0123abcd", "--scope", "email:read",
				"--name", "parallel", "--expires-in", "1h",
			})
		})
	}
	wg.Wait()
	close(results)
	write.Close()
	io.Copy(io.Discard, read)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	var count int
	err = db.Pool.QueryRow(t.Context(), "select count(*) from access_tokens where tenant='e2e/tokens-0123abcd'").Scan(&count)
	if err != nil || successes != 1 || count != 10 {
		t.Fatalf("concurrent creates: successes=%d stored=%d error=%v, want 1 and 10", successes, count, err)
	}
}
