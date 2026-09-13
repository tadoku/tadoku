package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

func TestImportPolicies(t *testing.T) {
	for _, tt := range []struct {
		name, dir, file, pkg, target, diagnostic string
	}{
		{"transport-app", "transport/http", "handler.go", "http", "services/tadoku-api/app", ""},
		{"transport-http-types", "transport/http", "handler.go", "http", "services/tadoku-api/generated/openapi", ""},
		{"transport-feature", "transport/http", "handler.go", "http", "services/tadoku-api/features/content", "import-denied"},
		{"transport-sql", "transport/http", "handler.go", "http", "services/tadoku-api/generated/sqlc/content", "import-denied"},
		{"transport-echo", "transport/http", "handler.go", "http", "github.com/labstack/echo/v4", "import-denied"},
		{"test-only-testify", "transport/http", "handler_test.go", "http", "github.com/stretchr/testify/require", "import-denied"},
		{"external-test-only-echo", "transport/http", "handler_test.go", "http_test", "github.com/labstack/echo/v4", "import-denied"},
		{"app-feature", "app", "app.go", "app", "services/tadoku-api/features/content", ""},
		{"app-transaction", "app", "app.go", "app", "services/tadoku-api/infra/postgres", ""},
		{"app-generated", "app", "app.go", "app", "services/tadoku-api/generated/openapi", "import-denied"},
		{"app-pgx", "app", "app.go", "app", "github.com/jackc/pgx/v5", "import-denied"},
		{"app-provider", "app", "app.go", "app", "github.com/ory/keto-client-go", "import-denied"},
		{"app-sql", "app", "app.go", "app", "database/sql", "import-denied"},
		{"feature-own-sql", "features/content", "repository.go", "content", "services/tadoku-api/generated/sqlc/content", ""},
		{"new-feature-own-sql", "features/profile", "repository.go", "profile", "services/tadoku-api/generated/sqlc/profile", ""},
		{"feature-pgx", "features/content", "repository.go", "content", "github.com/jackc/pgx/v5", ""},
		{"feature-other-sql", "features/content", "repository.go", "content", "services/tadoku-api/generated/sqlc/access", "import-denied"},
		{"feature-sideways", "features/content", "content.go", "content", "services/tadoku-api/features/access", "import-denied"},
		{"feature-app", "features/content", "content.go", "content", "services/tadoku-api/app", "import-denied"},
		{"feature-transport", "features/content", "content.go", "content", "services/tadoku-api/transport/http", "import-denied"},
		{"feature-http-types", "features/content", "content.go", "content", "services/tadoku-api/generated/openapi", "import-denied"},
		{"repository-fixture", "features/content", "repository_test.go", "content_test", "services/tadoku-api/internal/testpostgres", ""},
		{"repository-test-sideways", "features/content", "repository_test.go", "content_test", "services/tadoku-api/features/access", "import-denied"},
		{"pure-domain-uuid", "features/content/domain", "value.go", "domain", "github.com/google/uuid", ""},
		{"pure-domain-driver", "features/content/domain", "value.go", "domain", "github.com/jackc/pgx/v5", "import-denied"},
		{"pure-domain-generated", "features/content/domain", "value.go", "domain", "services/tadoku-api/generated/sqlc/content", "import-denied"},
		{"infrastructure-feature", "infra/postgres", "db.go", "postgres", "services/tadoku-api/features/content", "import-denied"},
		{"internal-feature", "internal/timex", "clock.go", "timex", "services/tadoku-api/features/content", "import-denied"},
		{"generated-feature", "generated/sqlc/content", "db.go", "content", "services/tadoku-api/features/content", "import-denied"},
		{"legacy-echo", "e2e/legacy", "content.go", "legacy", "github.com/labstack/echo/v4", ""},
		{"legacy-testify", "e2e/legacy", "content_test.go", "legacy_test", "github.com/stretchr/testify/require", "import-denied"},
		{"e2e-echo", "e2e", "endpoint_test.go", "e2e_test", "github.com/labstack/echo/v4", "import-denied"},
		{"uncovered", "unplanned", "file.go", "unplanned", "fmt", "uncovered-package"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := fixture(t)
			target := tt.target
			if strings.HasPrefix(target, "services/") {
				target = "github.com/tadoku/tadoku/" + target
			}
			// Inactive build tags must not hide imports from the architecture gate.
			writeFile(t, root, "services/tadoku-api/"+tt.dir+"/"+tt.file,
				fmt.Sprintf("//go:build policyfixture\n\npackage %s\nimport _ %q\n", tt.pkg, target))
			var output bytes.Buffer
			err := check(root, &output)
			if tt.diagnostic == "" {
				if err != nil {
					t.Fatalf("check: %v\n%s", err, &output)
				}
			} else if err == nil || !strings.Contains(output.String(), tt.diagnostic) || !strings.Contains(output.String(), tt.file) {
				t.Fatalf("wanted %s at %s; err=%v\n%s", tt.diagnostic, tt.file, err, &output)
			}
		})
	}
}

func TestConfigurationFailsClosed(t *testing.T) {
	for _, tt := range []struct{ name, path, content, want string }{
		{"missing", ".depolicy.yaml", "", ".depolicy.yaml"},
		{"invalid", ".depolicy.yaml", "version: 99\npolicies: []\n", "unsupported config version"},
		{"empty", ".depolicy.yaml", "\n", "configuration is empty"},
		{"no-policies", ".depolicy.yaml", "version: 1\npolicies: []\n", "native import policy failed"},
		{"ambiguous", ".depolicy.yaml", "version: 1\npolicies:\n- id: one\n  packages: [local:...]\n  imports: {default: allow}\n- id: two\n  packages: [local:...]\n  imports: {default: allow}\n", "native import policy failed"},
		{"missing-module", "go.mod", "", "go.mod"},
		{"wrong-module", "go.mod", "module example.com/other\n", "expected Tadoku go.mod"},
		{"nested-config", "services/tadoku-api/internal/.depolicy.yaml", "version: 1\npolicies: []\n", "nested configuration"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := fixture(t)
			if tt.content == "" {
				if err := os.Remove(filepath.Join(root, tt.path)); err != nil {
					t.Fatal(err)
				}
			} else {
				writeFile(t, root, tt.path, tt.content)
			}
			var output bytes.Buffer
			if err := check(root, &output); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("wanted %q; err=%v\n%s", tt.want, err, &output)
			}
		})
	}
}

func TestSourceScopeFailsClosed(t *testing.T) {
	for _, name := range []string{"missing-source", "missing-tests", "malformed-import", "symlink"} {
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			const source = "services/tadoku-api/internal/timex/clock.go"
			switch name {
			case "missing-source", "missing-tests":
				path := source
				if name == "missing-tests" {
					path = "services/tadoku-api/internal/timex/clock_test.go"
				}
				if err := os.Remove(filepath.Join(root, path)); err != nil {
					t.Fatal(err)
				}
			case "malformed-import":
				writeFile(t, root, source, "package timex\nimport (\n")
			case "symlink":
				if err := os.Symlink("clock.go", filepath.Join(root, "services/tadoku-api/internal/timex/linked.go")); err != nil {
					t.Fatal(err)
				}
			}
			var output bytes.Buffer
			if err := check(root, &output); err == nil {
				t.Fatalf("unsafe source scope passed: %s", &output)
			}
		})
	}
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	configPath, err := runfiles.Rlocation("_main/.depolicy.yaml")
	if err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, ".depolicy.yaml", string(config))
	writeFile(t, root, "go.mod", "module github.com/tadoku/tadoku\ngo 1.26.6\n")
	writeFile(t, root, "services/tadoku-api/internal/timex/clock.go", "package timex\nimport _ \"time\"\n")
	writeFile(t, root, "services/tadoku-api/internal/timex/clock_test.go", "package timex_test\nimport _ \"testing\"\n")
	return root
}

func writeFile(t *testing.T, root, path, content string) {
	t.Helper()
	path = filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
