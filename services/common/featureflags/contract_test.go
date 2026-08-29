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
	definition, ok := fixture.BooleanFlags[ReleaseLogEntryV2.Key()]
	require.True(t, ok)
	assert.Equal(t, ReleaseLogEntryV2.SafeDefault(), definition.SafeDefault)
}
