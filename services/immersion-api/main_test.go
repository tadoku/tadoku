package main

import (
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fliptclient "github.com/tadoku/tadoku/services/common/client/flipt"
	"github.com/valkey-io/valkey-go"
)

type stubValkeyClient struct {
	valkey.Client
}

func validConfig() Config {
	return Config{
		Port:          8080,
		JWKS:          "jwks",
		KratosURL:     "http://kratos",
		OathkeeperURL: "http://oathkeeper",
		KetoReadURL:   "http://keto-read",
		KetoWriteURL:  "http://keto-write",
		ValkeyURL:     "redis://valkey:6379",
		ValkeyTimeout: time.Second,
	}
}

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

func TestValkeyTimeoutDefaultsToOneSecond(t *testing.T) {
	cfg := Config{}

	require.NoError(t, envconfig.Process("IMMERSION_API_CONFIG_TEST", &cfg))
	assert.Equal(t, time.Second, cfg.ValkeyTimeout)
}

func TestConfigRejectsNonPositiveValkeyTimeout(t *testing.T) {
	cfg := validConfig()
	cfg.ValkeyTimeout = 0

	require.Error(t, validator.New().Struct(cfg))
}

func TestInitializeValkeyClientConfiguresBoundedStandaloneClient(t *testing.T) {
	client := &stubValkeyClient{}
	dialTimeout := 250 * time.Millisecond
	var captured valkey.ClientOption

	initialized, err := initializeValkeyClient(Config{
		ValkeyURL:     "redis://valkey:6379/3",
		ValkeyTimeout: dialTimeout,
	}, func(option valkey.ClientOption) (valkey.Client, error) {
		captured = option
		return client, nil
	})

	require.NoError(t, err)
	assert.Same(t, client, initialized)
	assert.Equal(t, []string{"valkey:6379"}, captured.InitAddress)
	assert.Equal(t, 3, captured.SelectDB)
	assert.Equal(t, dialTimeout, captured.Dialer.Timeout)
	assert.True(t, captured.ForceSingleClient)
	assert.True(t, captured.DisableRetry)
}

func TestInitializeValkeyClientContinuesWithClientAfterInitialDialFailure(t *testing.T) {
	client := &stubValkeyClient{}
	dialErr := errors.New("valkey unavailable")

	initialized, err := initializeValkeyClient(Config{
		ValkeyURL:     "redis://valkey:6379",
		ValkeyTimeout: time.Second,
	}, func(valkey.ClientOption) (valkey.Client, error) {
		return client, dialErr
	})

	require.NoError(t, err)
	assert.Same(t, client, initialized)
}

func TestInitializeValkeyClientRejectsInitialFailureWithoutClient(t *testing.T) {
	dialErr := errors.New("valkey unavailable")

	client, err := initializeValkeyClient(Config{
		ValkeyURL:     "redis://valkey:6379",
		ValkeyTimeout: time.Second,
	}, func(valkey.ClientOption) (valkey.Client, error) {
		return nil, dialErr
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, dialErr)
	assert.Nil(t, client)
}

func TestInitializeValkeyClientRejectsMalformedURL(t *testing.T) {
	called := false

	client, err := initializeValkeyClient(Config{
		ValkeyURL:     "://not-a-valkey-url",
		ValkeyTimeout: time.Second,
	}, func(valkey.ClientOption) (valkey.Client, error) {
		called = true
		return &stubValkeyClient{}, nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse valkey url")
	assert.Nil(t, client)
	assert.False(t, called)
}

func TestInitializeValkeyClientRejectsMultipleAddresses(t *testing.T) {
	called := false

	client, err := initializeValkeyClient(Config{
		ValkeyURL:     "redis://valkey-a:6379?addr=valkey-b:6379",
		ValkeyTimeout: time.Second,
	}, func(valkey.ClientOption) (valkey.Client, error) {
		called = true
		return &stubValkeyClient{}, nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one standalone address")
	assert.Nil(t, client)
	assert.False(t, called)
}
