package tenant_test

import (
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestParseTestTenant(t *testing.T) {
	for _, raw := range []string{
		"tadoku/branch-0123abcd",
		"e2e/a-00000000",
		"production-copy/" + strings.Repeat("a", 47) + "-abcdef12",
	} {
		key, err := tenant.ParseTestTenant(raw)
		if err != nil || key.String() != raw || key.Key().String() != raw {
			t.Fatalf("ParseTestTenant(%q): key=%q error=%v", raw, key.String(), err)
		}
	}

	for _, raw := range []string{
		"", "tadoku", "tadoku/prod", "branch-0123abcd", "tadoku/branch", "tadoku/branch-0123ABCd",
		"tadoku/-0123abcd", "tadoku/branch-0123abc", "tadoku/branch-0123abcde", "tadoku/branch_0123abcd",
		"tadoku/" + strings.Repeat("a", 48) + "-abcdef12", "tadoku/branch-0123abcd/extra",
	} {
		if _, err := tenant.ParseTestTenant(raw); err == nil {
			t.Errorf("accepted unsafe lifecycle key %q", raw)
		}
	}

	if (tenant.TestKey{}).Key() != (tenant.Key{}) {
		t.Fatal("zero test key became a parsed tenant")
	}
}
