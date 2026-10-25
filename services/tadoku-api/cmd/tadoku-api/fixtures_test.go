package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/common/postgresconfig"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
)

func validApplicationConfig(t *testing.T) config {
	t.Helper()
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	databaseURL, err := url.Parse(db.DSN)
	if err != nil {
		t.Fatal(err)
	}

	databasePort, err := strconv.Atoi(databaseURL.Port())
	if err != nil {
		t.Fatal(err)
	}

	upstream := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(upstream.Close)
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(jwks.Close)
	valkeyURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}

	return config{
		Port:                    0,
		MetricsPort:             0,
		ServiceName:             "tadoku-api-test",
		JWKS:                    jwks.URL,
		OathkeeperAuthzToken:    "callback-token",
		OathkeeperURL:           upstream.URL,
		ServiceAccountTokenPath: "/var/run/secrets/tokens/token",
		FliptURL:                upstream.URL + "/flipt",
		FliptEnvironment:        "local",
		FliptTestEnvironment:    "test",
		FliptNamespace:          "default",
		FliptUpdateInterval:     time.Minute,
		FliptRequestTimeout:     time.Second,
		FliptStartupTimeout:     10 * time.Millisecond,

		DialTimeout:           10 * time.Second,
		MaxTokenAge:           24 * time.Hour,
		ResponseHeaderTimeout: time.Second,
		RequestTimeout:        time.Second,
		IdleTimeout:           time.Second,
		ShutdownTimeout:       time.Second,

		PostgresMaxConnections: 4,
		Postgres: postgresconfig.Config{
			Host:     "127.0.0.1",
			Port:     uint16(databasePort),
			Database: databaseURL.Path[1:],
			User:     "postgres",
			Password: postgresconfig.NewSecret("postgres"),
			SSLMode:  "disable",
		},
		ValkeyURL:     valkeyURL,
		ValkeyTimeout: time.Second,
	}
}
