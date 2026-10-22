package e2e_test

import (
	"context"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type credentialCounts struct {
	expiredSessions, retainedSessions int
	expiredTokens, retainedTokens     int
	retainedFingerprint               string
}

type credentialExpiryTracer struct {
	cancelTenant string
	cancel       context.CancelFunc
	claimed      chan struct{}

	mu       sync.Mutex
	once     sync.Once
	canceled bool
	tenants  []string
}

type credentialBatchKey struct{}

func (*credentialExpiryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	_ pgx.TraceQueryStartData,
) context.Context {
	return ctx
}

func (*credentialExpiryTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (tracer *credentialExpiryTracer) TraceBatchStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceBatchStartData,
) context.Context {
	for _, query := range data.Batch.QueuedQueries {
		if strings.HasPrefix(query.SQL, "-- name: Claim :") {
			tracer.once.Do(func() { close(tracer.claimed) })
		}
		if !strings.HasPrefix(query.SQL, "-- name: DeleteExpiredCredentials :") {
			continue
		}

		key, _ := tenant.FromContext(ctx)
		tracer.mu.Lock()
		tracer.tenants = append(tracer.tenants, key.String())
		tracer.mu.Unlock()
		if key.String() == tracer.cancelTenant {
			return context.WithValue(ctx, credentialBatchKey{}, true)
		}
	}
	return ctx
}

func (*credentialExpiryTracer) TraceBatchQuery(context.Context, *pgx.Conn, pgx.TraceBatchQueryData) {}

func (tracer *credentialExpiryTracer) TraceBatchEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceBatchEndData) {
	if ctx.Value(credentialBatchKey{}) == nil {
		return
	}
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	if !tracer.canceled {
		tracer.canceled = true
		tracer.cancel()
	}
}

func (tracer *credentialExpiryTracer) scopedTenants() []string {
	tracer.mu.Lock()
	defer tracer.mu.Unlock()
	tenants := slices.Clone(tracer.tenants)
	slices.Sort(tenants)
	return slices.Compact(tenants)
}

// runCredentialExpiry runs the worker's startup maintenance at the fixture instant through the restricted
// application role. A non-empty cancelTenant cancels the worker after that tenant's first deletion batch commits.
func runCredentialExpiry(t *testing.T, scope jobqueue.Scope, cancelTenant string) []string {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	tracer := &credentialExpiryTracer{cancelTenant: cancelTenant, cancel: cancel, claimed: make(chan struct{})}
	config := api.db.AppPool.Config()
	config.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	application, err := worker.NewApplication(
		jobqueue.NewService(jobqueue.NewRepository(pool)),
		api.leaderboard,
		worker.Config{
			Identity:        identity.NewExpiry(identity.NewRepository(pool)),
			Scope:           scope,
			Concurrency:     1,
			ShutdownTimeout: time.Second,
			Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	timex.TheWorld(fixtureInstant, func() {
		done := make(chan error, 1)
		go func() { done <- application.Run(ctx) }()
		select {
		case <-tracer.claimed:
		case <-done:
			return
		case <-time.After(20 * time.Second):
			t.Error("worker did not finish startup maintenance")
		}
		cancel()
		if err := <-done; err != nil {
			t.Error(err)
		}
	})
	return tracer.scopedTenants()
}

func seedCredentialExpiry(t *testing.T) {
	t.Helper()
	_, err := api.db.Pool.Exec(t.Context(), `with owners (tenant, user_id, expired) as (
			values
				('tadoku/identity-a-0123abcd', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb'::uuid, 250),
				('tadoku/identity-b-0123abcd', 'cccccccc-cccc-4ccc-8ccc-cccccccccccc'::uuid, 120),
				('tadoku/prod', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa'::uuid, 130)
		), expired_sessions as (
			insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at)
			select tenant, 'expired-session-' || n, user_id, $1::timestamptz - interval '30 days',
				$1::timestamptz - n * interval '1 minute'
			from owners
			cross join lateral generate_series(1, expired) as n
		), retained_sessions as (
			insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at)
			select tenant, 'retained-session-' || label, user_id, $1::timestamptz - interval '1 day', $1::timestamptz + age
			from owners
			cross join (
				values
					('boundary', interval '0'),
					('next-second', interval '1 second'),
					('next-month', interval '30 days')
			) as retained (label, age)
		), expired_tokens as (
			insert into recovery_tokens (tenant, token_hash, user_id, expires_at, used_at)
			select tenant, 'expired-token-' || n, user_id, $1::timestamptz - n * interval '1 minute',
				case
					when n % 2 = 0 then $1::timestamptz - interval '2 hours'
				end
			from owners
			cross join lateral generate_series(1, expired) as n
		)
		insert into recovery_tokens (tenant, token_hash, user_id, expires_at, used_at)
		select tenant, 'retained-token-' || label, user_id, $1::timestamptz + age, used_at
		from owners
		cross join (
			values
				('boundary', interval '0', null::timestamptz),
				('unused', interval '30 minutes', null),
				('consumed', interval '30 minutes', $1::timestamptz - interval '10 minutes'),
				('invalidated', interval '1 minute', $1::timestamptz - interval '1 minute')
		) as retained (label, age, used_at)`,
		fixtureInstant,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func credentialExpiryCounts(t *testing.T) map[string]credentialCounts {
	t.Helper()
	rows, err := api.db.Pool.Query(t.Context(), `select
			t.key,
			(select count(*) from sessions s where s.tenant = t.key and s.expires_at < $1),
			(select count(*) from sessions s where s.tenant = t.key and s.expires_at >= $1),
			(select count(*) from recovery_tokens r where r.tenant = t.key and r.expires_at < $1),
			(select count(*) from recovery_tokens r where r.tenant = t.key and r.expires_at >= $1),
			md5(
				coalesce((
					select string_agg(s.token_hash || '@' || s.expires_at, ',' order by s.token_hash)
					from sessions s
					where s.tenant = t.key and s.expires_at >= $1
				), '') || '|' || coalesce((
					select string_agg(r.token_hash || '@' || r.expires_at || '@' || coalesce(r.used_at::text, ''), ','
						order by r.token_hash)
					from recovery_tokens r
					where r.tenant = t.key and r.expires_at >= $1
				), '')
			)
		from tenants t
		where t.key in ('tadoku/identity-a-0123abcd', 'tadoku/identity-b-0123abcd', 'tadoku/prod')`,
		fixtureInstant,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	counts := make(map[string]credentialCounts)
	for rows.Next() {
		var key string
		var row credentialCounts
		err := rows.Scan(
			&key,
			&row.expiredSessions,
			&row.retainedSessions,
			&row.expiredTokens,
			&row.retainedTokens,
			&row.retainedFingerprint,
		)
		if err != nil {
			t.Fatal(err)
		}
		counts[key] = row
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return counts
}

func requireCredentialCounts(
	t *testing.T,
	step string,
	got, before map[string]credentialCounts,
	expired map[string][2]int,
) {
	t.Helper()
	for key, want := range expired {
		row := got[key]
		if row.expiredSessions != want[0] || row.expiredTokens != want[1] {
			t.Errorf(
				"%s: %s expired sessions=%d recovery tokens=%d; want %d and %d",
				step, key, row.expiredSessions, row.expiredTokens, want[0], want[1],
			)
		}
		if row.retainedSessions != before[key].retainedSessions ||
			row.retainedTokens != before[key].retainedTokens ||
			row.retainedFingerprint != before[key].retainedFingerprint {
			t.Errorf(
				"%s: %s retained sessions=%d recovery tokens=%d changed from %d and %d or their rows changed",
				step, key, row.retainedSessions, row.retainedTokens,
				before[key].retainedSessions, before[key].retainedTokens,
			)
		}
	}
}

func TestIdentityWorkerRemovesOnlyExpiredCredentials(t *testing.T) {
	resetIdentity(t)
	cookie := identityLoginCookie(t)
	validToken := seedRecoveryToken(t, 'v', fixtureInstant.Add(time.Hour), nil)
	expiredToken := seedRecoveryToken(t, 'e', fixtureInstant.Add(-time.Second), nil)
	seedCredentialExpiry(t)
	branchKey, err := tenant.Parse(identityBranchB)
	if err != nil {
		t.Fatal(err)
	}
	branchScope, err := jobqueue.OnlyTenant(branchKey)
	if err != nil {
		t.Fatal(err)
	}
	baseScope, err := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
	if err != nil {
		t.Fatal(err)
	}

	before := credentialExpiryCounts(t)
	requireCredentialCounts(t, "seed", before, before, map[string][2]int{
		identityBranchA: {250, 251},
		identityBranchB: {120, 120},
		"tadoku/prod":   {130, 130},
	})
	if before[identityBranchA].retainedSessions != 4 || before[identityBranchA].retainedTokens != 5 ||
		before["tadoku/prod"].retainedSessions != 3 || before["tadoku/prod"].retainedTokens != 4 {
		t.Fatalf("retained fixture rows = %+v", before)
	}

	tenants := runCredentialExpiry(t, branchScope, "")
	if !slices.Equal(tenants, []string{identityBranchB}) {
		t.Errorf("branch worker scoped credential cleanup to %v; want only %s", tenants, identityBranchB)
	}
	requireCredentialCounts(t, "branch worker", credentialExpiryCounts(t), before, map[string][2]int{
		identityBranchA: {250, 251},
		identityBranchB: {0, 0},
		"tadoku/prod":   {130, 130},
	})

	runCredentialExpiry(t, baseScope, identityBranchA)
	requireCredentialCounts(t, "canceled base worker", credentialExpiryCounts(t), before, map[string][2]int{
		identityBranchA: {150, 151},
		identityBranchB: {0, 0},
		"tadoku/prod":   {130, 130},
	})

	tenants = runCredentialExpiry(t, baseScope, "")
	for _, key := range []string{identityBranchA, identityBranchB, "tadoku/prod"} {
		if !slices.Contains(tenants, key) {
			t.Errorf("base worker credential cleanup tenants %v omit %s", tenants, key)
		}
	}
	removed := credentialExpiryCounts(t)
	requireCredentialCounts(t, "base worker", removed, before, map[string][2]int{
		identityBranchA: {0, 0},
		identityBranchB: {0, 0},
		"tadoku/prod":   {0, 0},
	})

	runCredentialExpiry(t, baseScope, "")
	if repeated := credentialExpiryCounts(t); !maps.Equal(repeated, removed) {
		t.Errorf("repeated cleanup changed credentials: %+v; want %+v", repeated, removed)
	}

	response := identityRequest(t, api.handler, identityBranchA, "GET", "/identity/session", "", cookie, nil)
	if identityView(t, response).Account == nil {
		t.Error("cleanup revoked an active session")
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/recovery/"+expiredToken, "", "", nil)
	requireIdentityStatus(t, response, http.StatusBadRequest)
	if !strings.Contains(response.Body.String(), `"code":"invalid_recovery_token"`) {
		t.Errorf("removed expired recovery token answered %s", response.Body)
	}
	response = identityRequest(t, api.handler, identityBranchA, "GET", "/identity/recovery/"+validToken, "", "", nil)
	requireIdentityStatus(t, response, http.StatusSeeOther)
}
