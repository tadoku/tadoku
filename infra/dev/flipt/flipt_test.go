package flipt_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type featureContract struct {
	BooleanFlags map[string]struct {
		SafeDefault bool `json:"safeDefault"`
	} `json:"booleanFlags"`
}

func TestSeedMatchesApplicationContract(t *testing.T) {
	contractContents := readRunfile(t, "feature-flags.contract.json")
	var contract featureContract
	require.NoError(t, json.Unmarshal(contractContents, &contract))

	seed := string(readRunfile(t, "infra/dev/flipt/seed/default/features.yaml"))
	assert.Contains(t, seed, "version: \"1.4\"")
	assert.Contains(t, seed, "namespace:\n  key: default")
	assert.Equal(t, len(contract.BooleanFlags), strings.Count(seed, "\n  - key: "))

	for key, definition := range contract.BooleanFlags {
		pattern := regexp.MustCompile(
			`(?m)^  - key: ` + regexp.QuoteMeta(key) +
				`\n(?:    [^\n]+\n)*?    type: BOOLEAN_FLAG_TYPE` +
				`\n(?:    [^\n]+\n)*?    enabled: (true|false)$`,
		)
		match := pattern.FindStringSubmatch(seed)
		require.Len(t, match, 2, "seed must define boolean contract flag %q", key)
		assert.Equal(t, definition.SafeDefault, match[1] == "true")
	}
}

func TestDeploymentBootstrapsRestrictedDisposableGitRepository(t *testing.T) {
	manifest := string(readRunfile(t, "infra/dev/flipt/flipt.yaml"))

	assert.Contains(t, manifest, "tadoku.dev/flipt-seed-sha256: __FLIPT_SEED_SHA256__")
	assert.Contains(t, manifest, "image: docker.io/alpine/git:2.47.2@sha256:")
	assert.Contains(t, manifest, "git init --initial-branch=main")
	assert.Contains(t, manifest, "chmod 0644 /var/lib/flipt/default/features.yaml")
	assert.Contains(t, manifest, "git config --global --add safe.directory /var/lib/flipt")
	assert.Contains(t, manifest, "git add default/features.yaml")
	assert.Contains(t, manifest, "name: seed\n              mountPath: /seed\n              readOnly: true")
	assert.Contains(t, manifest, "name: seed\n          configMap:\n            name: flipt-seed")
	assert.Contains(t, manifest, "name: repository\n          emptyDir:\n            medium: Memory")
	assert.Contains(t, manifest, "name: repository\n              mountPath: /var/lib/flipt")
	assert.Contains(t, manifest, "allowPrivilegeEscalation: false")
	assert.Contains(t, manifest, "readOnlyRootFilesystem: true")
	assert.Contains(t, manifest, "runAsNonRoot: true")
	assert.Contains(t, manifest, "seccompProfile:\n              type: RuntimeDefault")
	assert.NotContains(t, manifest, "persistentVolumeClaim:")
}

func TestFliptUsesLocalSeededStorageWithoutRemote(t *testing.T) {
	manifest := string(readRunfile(t, "infra/dev/flipt/flipt.yaml"))
	assert.Contains(t, manifest, "type: local\n          path: /var/lib/flipt")
	assert.Contains(t, manifest, "branch: main")
	assert.Contains(t, manifest, "directory: default")
	assert.NotContains(t, manifest, "remote:")

	tiltfile := string(readRunfile(t, "infra/dev/flipt/Tiltfile"))
	assert.Contains(t, tiltfile, "read_file(seed_path)")
	assert.Contains(t, tiltfile, "__FLIPT_SEED_SHA256__")
	assert.Contains(t, tiltfile, "binaryData:")

	seed := readRunfile(t, "infra/dev/flipt/seed/default/features.yaml")
	digest := sha256.Sum256(seed)
	assert.Len(t, hex.EncodeToString(digest[:]), 64)
}

func readRunfile(t *testing.T, path string) []byte {
	t.Helper()
	resolved, err := bazel.Runfile(path)
	require.NoError(t, err)
	contents, err := os.ReadFile(resolved)
	require.NoError(t, err)
	return contents
}
