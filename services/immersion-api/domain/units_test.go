package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
)

func TestUnitDefinitions(t *testing.T) {
	definitions := domain.UnitDefinitions()

	assert.Len(t, definitions, 13)
	assert.Equal(t, domain.UnitKeyReadingPage, definitions[0].Key)

	definitions[0].Name = "changed"
	resolved, ok := domain.UnitDefinitionByKey(domain.UnitKeyReadingPage)
	require.True(t, ok)
	assert.Equal(t, "Page", resolved.Name)
	assert.Equal(t, int32(1), resolved.ActivityID)
}

func TestUnitDefinitionByKeyRejectsUnknownKey(t *testing.T) {
	_, ok := domain.UnitDefinitionByKey("unknown")

	assert.False(t, ok)
}
