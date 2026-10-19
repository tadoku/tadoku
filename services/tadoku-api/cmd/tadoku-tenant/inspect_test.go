package main

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

func TestInspectTestTenant(t *testing.T) {
	accessTokenDatabase(t)
	for _, c := range []struct {
		key    string
		exists bool
	}{
		{"e2e/tokens-0123abcd", true},
		{"e2e/missing-0123abcd", false},
	} {
		read, write, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		old := os.Stdout
		os.Stdout = write
		err = run(t.Context(), []string{"inspect", "--tenant", c.key})
		os.Stdout = old
		write.Close()
		raw, readErr := io.ReadAll(read)
		read.Close()
		if err != nil || readErr != nil {
			t.Fatalf("inspect tenant: %v, %v", err, readErr)
		}
		var result struct {
			Tenant string
			Exists bool
		}
		if json.Unmarshal(raw, &result) != nil || result.Tenant != c.key || result.Exists != c.exists {
			t.Fatalf("inspect result %s", raw)
		}
	}
	if err := run(t.Context(), []string{"inspect", "--tenant", "tadoku/prod"}); err == nil {
		t.Fatal("inspect accepted canonical tenant")
	}
}
