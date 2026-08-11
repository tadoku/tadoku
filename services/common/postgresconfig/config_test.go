package postgresconfig

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setIndividual(t *testing.T) {
	t.Helper()
	t.Setenv("TEST_HOST", "2001:db8::1")
	t.Setenv("TEST_DATABASE", "tadoku")
	t.Setenv("TEST_USER", "user@name")
	t.Setenv("TEST_PASSWORD", "sentinel:/?#[]@!$&'()*+,;=")
	t.Setenv("TEST_SSLMODE", "verify-full")
}

func TestLoadIndividualAndConnConfig(t *testing.T) {
	setIndividual(t)
	cfg, err := Load("TEST", "TEST_URL")
	require.NoError(t, err)
	assert.Equal(t, uint16(5432), cfg.Port)
	parsed, err := cfg.ConnConfig()
	require.NoError(t, err)
	assert.Equal(t, "2001:db8::1", parsed.Host)
	assert.Equal(t, "sentinel:/?#[]@!$&'()*+,;=", parsed.Password)
	assert.NotContains(t, fmt.Sprint(cfg), "sentinel")
}

func TestLoadRejectsPartialMixedAndInvalid(t *testing.T) {
	t.Run("partial", func(t *testing.T) {
		t.Setenv("TEST_HOST", "db")
		_, err := Load("TEST", "TEST_URL")
		assert.ErrorContains(t, err, "missing postgres configuration")
	})
	t.Run("mixed", func(t *testing.T) {
		setIndividual(t)
		t.Setenv("TEST_URL", "postgres://legacy")
		_, err := Load("TEST", "TEST_URL")
		assert.ErrorContains(t, err, "no longer supported")
	})
	t.Run("port", func(t *testing.T) {
		setIndividual(t)
		t.Setenv("TEST_PORT", "65536")
		_, err := Load("TEST", "TEST_URL")
		assert.ErrorContains(t, err, "between 1 and 65535")
	})
	t.Run("sslmode", func(t *testing.T) {
		setIndividual(t)
		t.Setenv("TEST_SSLMODE", "surprise")
		_, err := Load("TEST", "TEST_URL")
		assert.ErrorContains(t, err, "SSLMODE is invalid")
	})
}

func TestLegacyIsRejected(t *testing.T) {
	const legacy = "postgres://user:sentinel-secret@db/database?sslmode=require"
	t.Setenv("TEST_URL", legacy)
	_, err := Load("TEST", "TEST_URL")
	assert.ErrorContains(t, err, "no longer supported")
}
