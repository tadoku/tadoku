package featureflags

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contractFixture struct {
	BooleanFlags map[string]struct {
		SafeDefault bool `json:"safeDefault"`
	} `json:"booleanFlags"`
}

func TestRegistryMatchesCrossStackContract(t *testing.T) {
	path, err := bazel.Runfile("feature-flags.contract.json")
	require.NoError(t, err)
	contents, err := os.ReadFile(path)
	require.NoError(t, err)

	var fixture contractFixture
	require.NoError(t, json.Unmarshal(contents, &fixture))
	require.Len(t, fixture.BooleanFlags, 1)
	for _, flag := range []BooleanFlag{ReleaseLogEntryV2} {
		definition, ok := fixture.BooleanFlags[flag.Key()]
		require.True(t, ok)
		assert.Equal(t, flag.SafeDefault(), definition.SafeDefault)
	}
}
