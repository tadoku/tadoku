package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fliptclient "github.com/tadoku/tadoku/services/common/client/flipt"
)

func TestDisabledFliptDoesNotInitializeProvider(t *testing.T) {
	called := false
	cfg := Config{FliptEnabled: false}

	provider, err := initializeFeatureFlagProvider(cfg, func() (*fliptclient.Client, error) {
		called = true
		return nil, errors.New("must not initialize")
	})

	require.NoError(t, err)
	assert.Nil(t, provider)
	assert.False(t, called)
}
