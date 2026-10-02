package postgresconfig

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
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

func TestLoadIndividualAndURL(t *testing.T) {
	setIndividual(t)
	cfg, err := Load("TEST", "TEST_URL")
	require.NoError(t, err)
	assert.Equal(t, uint16(5432), cfg.Port)
	parsed, err := url.Parse(cfg.URL().Reveal())
	require.NoError(t, err)
	assert.Equal(t, "2001:db8::1", parsed.Hostname())
	password, ok := parsed.User.Password()
	assert.True(t, ok)
	assert.Equal(t, "sentinel:/?#[]@!$&'()*+,;=", password)
	assert.Empty(t, parsed.Query().Get("application_name"))
	assert.NotContains(t, fmt.Sprint(cfg), "sentinel")
}

func TestWithApplicationNameLabelsConnections(t *testing.T) {
	setIndividual(t)
	cfg, err := Load("TEST", "TEST_URL")
	require.NoError(t, err)

	labeled := cfg.WithApplicationName(" tadoku-api ")
	assert.Empty(t, cfg.ApplicationName)
	assert.Equal(t, "tadoku-api", labeled.ApplicationName)
	assert.Contains(t, labeled.URL().Reveal(), "application_name=tadoku-api")
	assert.NotContains(t, cfg.URL().Reveal(), "application_name=")

	parsed, err := url.Parse(labeled.URL().Reveal())
	require.NoError(t, err)
	assert.Equal(t, "tadoku-api", parsed.Query().Get("application_name"))
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

func TestSecretsAreRedactedWhenFormattedOrLogged(t *testing.T) {
	const sentinel = "sentinel-password"
	setIndividual(t)
	t.Setenv("TEST_PASSWORD", sentinel)
	cfg, err := Load("TEST", "TEST_URL")
	if err != nil {
		t.Fatal(err)
	}
	exposes := func(output string) bool {
		return strings.Contains(output, sentinel) || strings.Contains(output, hex.EncodeToString([]byte(sentinel)))
	}

	values := []struct {
		name  string
		value any
	}{{"config", cfg}, {"password", cfg.Password}, {"dsn", cfg.URL()}}
	for _, value := range values {
		for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%d"} {
			if output := fmt.Sprintf(verb, value.value); exposes(output) {
				t.Errorf("%s formatted with %s exposed the secret: %s", value.name, verb, output)
			}
		}
		encoded, err := json.Marshal(value.value)
		if err != nil {
			t.Fatal(err)
		}
		if exposes(string(encoded)) {
			t.Errorf("%s marshaled to JSON exposed the secret: %s", value.name, encoded)
		}
		for _, handler := range []func(io.Writer) slog.Handler{
			func(w io.Writer) slog.Handler { return slog.NewTextHandler(w, nil) },
			func(w io.Writer) slog.Handler { return slog.NewJSONHandler(w, nil) },
		} {
			var logs bytes.Buffer
			slog.New(handler(&logs)).Info("postgres", value.name, value.value)
			if exposes(logs.String()) {
				t.Errorf("%s logged exposed the secret: %s", value.name, logs.String())
			}
		}
	}
}

func TestLegacyIsRejected(t *testing.T) {
	const legacy = "postgres://user:sentinel-secret@db/database?sslmode=require"
	t.Setenv("TEST_URL", legacy)
	_, err := Load("TEST", "TEST_URL")
	assert.ErrorContains(t, err, "no longer supported")
}
