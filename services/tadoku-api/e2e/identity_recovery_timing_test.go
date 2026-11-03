package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

type recoveryWriteDelay struct {
	delay  time.Duration
	writes int
}

func (d *recoveryWriteDelay) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	if strings.HasPrefix(data.SQL, "-- name: CreateRecoveryToken :") {
		d.writes++
		time.Sleep(d.delay)
	}
	return ctx
}

func (*recoveryWriteDelay) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func recoveryTimingHandler(t *testing.T, delay *recoveryWriteDelay) http.Handler {
	t.Helper()
	config := api.db.AppPool.Config()
	config.ConnConfig.Tracer = delay
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	option, err := valkeygo.ParseURL(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	// Test safety: use the suite's leased DB, but keep the client inside the synctest bubble.
	option.SelectDB = leaderboardValkeyDatabase()
	option.ForceSingleClient = true
	option.DisableRetry = true
	option.DisableAutoPipelining = true
	client, err := valkeygo.NewClient(option)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)

	jwks, err := os.ReadFile("testdata/authentication.jwks.json")
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Connection", "close")
		_, _ = w.Write(jwks)
	}))
	t.Cleanup(server.Close)

	handler, _, err := newTestRouterWithLogger(
		ctx, pool, pool, slog.New(slog.NewTextHandler(io.Discard, nil)), withIdentityValkey(client),
		func(endpoints *routerEndpoints) { endpoints.jwks = server.URL },
	)
	if err != nil {
		t.Fatal(err)
	}
	server.Close()
	return handler
}

func TestIdentityRecoveryProcessingTime(t *testing.T) {
	for _, tc := range []struct {
		name  string
		delay time.Duration
		want  time.Duration
	}{
		{"below padding floor", 100 * time.Millisecond, 300 * time.Millisecond},
		{"above padding floor", 800 * time.Millisecond, 800 * time.Millisecond},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetIdentity(t)
			synctest.Test(t, func(t *testing.T) {
				delay := &recoveryWriteDelay{delay: tc.delay}
				handler := recoveryTimingHandler(t, delay)
				for _, email := range []string{"reader@example.test", "unknown@example.test"} {
					request := newIdentityRequest(
						t, identityBranchA, "POST", "/identity/recovery",
						`{"email":"`+email+`","return_to":"`+recoverySettingsURL+`"}`, "", nil,
					)
					response := httptest.NewRecorder()
					var elapsed time.Duration
					atFixtureInstant(func() {
						started := time.Now()
						handler.ServeHTTP(response, request)
						elapsed = time.Since(started)
					})
					requireRecoveryAccepted(t, response)

					want := 300 * time.Millisecond
					if email == "reader@example.test" {
						want = tc.want
					}
					if elapsed != want {
						t.Errorf("%s elapsed=%s, want %s", email, elapsed, want)
					}
					t.Logf("%s processing delay=%s elapsed=%s", email, tc.delay, elapsed)
				}
				if delay.writes != 1 {
					t.Errorf("delayed %d recovery writes, want one known-account write", delay.writes)
				}
			})
			requireRecoveryWrites(t, identityBranchA, 1)
		})
	}
}

func TestIdentityRecoveryMinimumDelay(t *testing.T) {
	resetIdentity(t)
	for _, email := range []string{"reader@example.test", "unknown@example.test"} {
		request := newIdentityRequest(
			t, identityBranchA, "POST", "/identity/recovery",
			`{"email":"`+email+`","return_to":"`+recoverySettingsURL+`"}`, "", nil,
		)
		response := httptest.NewRecorder()
		atFixtureInstant(func() {
			started := time.Now()
			api.handler.ServeHTTP(response, request)
			elapsed := time.Since(started)
			if elapsed < 300*time.Millisecond {
				t.Errorf("%s elapsed=%s, want at least 300ms", email, elapsed)
			}
			t.Logf("real-clock %s elapsed=%s", email, elapsed)
		})
		requireRecoveryAccepted(t, response)
	}
	requireRecoveryWrites(t, identityBranchA, 1)
}
