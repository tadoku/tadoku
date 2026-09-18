package testvalkey

import (
	"strings"
	"testing"
)

func TestURLRejectsCredentialsAndNonLoopbackHosts(t *testing.T) {
	for _, raw := range []string{
		"redis://user:password@127.0.0.1:6379",
		"redis://valkey.test:6379",
		"redis://127.0.0.1:6379/1",
		"redis://127.0.0.1:6379?db=1",
	} {
		t.Run(strings.ReplaceAll(raw, "/", "_"), func(t *testing.T) {
			t.Setenv("TADOKU_TEST_VALKEY_URL", raw)
			if _, err := URL(); err == nil {
				t.Errorf("URL accepted %q", raw)
			}
		})
	}
}
