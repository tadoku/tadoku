package e2e_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/app/accountdeletion"
	"github.com/tadoku/tadoku/services/tadoku-api/app/permissionmodel"
	"github.com/tadoku/tadoku/services/tadoku-api/app/worker"
	"github.com/tadoku/tadoku/services/tadoku-api/domain/jobs"
	featureaudit "github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	featureauthz "github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	featureprofile "github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

const (
	accountDeletionDir = "testdata/AccountDeletion"
	readerUserID       = "11111111-1111-4111-8111-111111111111"
	adminUserID        = "22222222-2222-4222-8222-222222222222"
	bannedUserID       = "33333333-3333-4333-8333-333333333333"
	bystanderUserID    = "44444444-4444-4444-8444-444444444444"
)

const othersSnapshotSQL = `select jsonb_build_object(
	'users', (select jsonb_agg(to_jsonb(u) order by u.tenant, u.id) from users u
		where not (u.tenant = 'tadoku/prod' and u.id = $1)),
	'accounts', (select jsonb_agg(to_jsonb(a) order by a.tenant, a.user_id) from accounts a
		where not (a.tenant = 'tadoku/prod' and a.user_id = $1)),
	'sessions', (select jsonb_agg(to_jsonb(s) order by s.tenant, s.token_hash) from sessions s
		where not (s.tenant = 'tadoku/prod' and s.user_id = $1)),
	'logs', (select jsonb_agg(to_jsonb(l) order by l.tenant, l.id) from logs l
		where not (l.tenant = 'tadoku/prod' and l.user_id = $1)),
	'contest_logs', (select jsonb_agg(to_jsonb(cl) order by cl.tenant, cl.contest_id, cl.log_id) from contest_logs cl
		join logs l on l.tenant = cl.tenant and l.id = cl.log_id
		where not (l.tenant = 'tadoku/prod' and l.user_id = $1)),
	'log_tags', (select jsonb_agg(to_jsonb(t) order by t.tenant, t.log_id, t.tag) from log_tags t
		where not (t.tenant = 'tadoku/prod' and t.user_id = $1)),
	'registrations', (select jsonb_agg(to_jsonb(r) order by r.tenant, r.id) from contest_registrations r
		where not (r.tenant = 'tadoku/prod' and r.user_id = $1)),
	'contests', (select jsonb_agg(to_jsonb(c) order by c.tenant, c.id) from contests c
		where not (c.tenant = 'tadoku/prod' and c.owner_user_id = $1)),
	'relationships', (select jsonb_agg(to_jsonb(p) order by p.tenant, p.object_type, p.object_id, p.relation)
		from permission_relationships p
		where not (p.tenant = 'tadoku/prod' and p.subject_type = 'user' and p.subject_id = $1::text))
)::text`

const targetSnapshotSQL = `select jsonb_build_object(
	'users', (select jsonb_agg(to_jsonb(u)) from users u where u.tenant = 'tadoku/prod' and u.id = $1),
	'accounts', (select count(*) from accounts where tenant = 'tadoku/prod' and user_id = $1),
	'sessions', (select count(*) from sessions where tenant = 'tadoku/prod' and user_id = $1),
	'logs', (select jsonb_agg(to_jsonb(l) order by l.id) from logs l where l.tenant = 'tadoku/prod' and l.user_id = $1),
	'contest_logs', (select count(*) from contest_logs cl
		join logs l on l.tenant = cl.tenant and l.id = cl.log_id
		where l.tenant = 'tadoku/prod' and l.user_id = $1),
	'registrations', (select count(*) from contest_registrations where tenant = 'tadoku/prod' and user_id = $1),
	'audit', (select count(*) from moderation_audit_log),
	'emails', (select count(*) from emails)
)::text`

const finishedHistorySQL = `select jsonb_agg(jsonb_build_object(
		'contest_id', cl.contest_id,
		'log_id', cl.log_id,
		'user_id', l.user_id,
		'language_code', l.language_code,
		'duration_seconds', cl.duration_seconds,
		'score', cl.computed_score,
		'eligible_official', l.eligible_official_leaderboard
	) order by cl.contest_id, cl.log_id)::text
	from contest_logs cl
	join contests c on c.tenant = cl.tenant and c.id = cl.contest_id
	join logs l on l.tenant = cl.tenant and l.id = cl.log_id
	where cl.tenant = 'tadoku/prod'
		and c.contest_end < '2026-09-12'
		and l.deleted_at is null`

func resetAccountDeletion(t *testing.T) {
	t.Helper()
	api.reset(t, "")
	if err := api.db.Reset(t.Context(), accountDeletionDir+"/setup.sql", "testdata/accounts.sql"); err != nil {
		t.Fatal(err)
	}
	err := resetRelationships(
		t.Context(),
		api.db.Pool,
		journeysDir+"/relationships.json",
		accountDeletionDir+"/relationships.json",
	)
	if err != nil {
		t.Fatal(err)
	}
}

func fileOperatorDeletion(t *testing.T, userID, operatorID, reason string) (int64, error) {
	t.Helper()
	model, err := permissionmodel.New()
	if err != nil {
		t.Fatal(err)
	}
	pool := api.db.AppPool
	tuples := featureauthz.NewRepository(pool)
	filing := accountdeletion.NewFiling(
		featureauthz.NewRoles(permissions.NewChecker(tuples, model), permissions.NewRelationships(tuples, model)),
		featureprofile.NewService(featureprofile.NewRepository(pool), identity.NewReader(pool)),
		featureaudit.NewService(featureaudit.NewRepository(pool)),
		jobqueue.NewService(jobqueue.NewRepository(pool)),
	)
	request, err := accountdeletion.OperatorRequest(uuid.MustParse(userID), uuid.MustParse(operatorID), reason)
	if err != nil {
		t.Fatal(err)
	}

	var jobID int64
	atFixtureInstant(func() {
		ctx := tenant.WithKey(t.Context(), tenant.Production())
		jobID, err = worker.FileAccountDeletion(ctx, pool, filing, request)
	})
	return jobID, err
}

func enqueueAccountDeletionJob(t *testing.T, userID string) {
	t.Helper()
	queue := jobqueue.NewService(jobqueue.NewRepository(api.db.AppPool))
	var err error
	atFixtureInstant(func() {
		ctx := tenant.WithKey(t.Context(), tenant.Production())
		err = queue.Enqueue(ctx, jobs.DeleteAccountV1{UserID: uuid.MustParse(userID)})
	})
	if err != nil {
		t.Fatal(err)
	}
}

// runDeletionWorker runs the worker until no job is pending or running, and tolerates failed jobs so that
// callers can assert them.
func runDeletionWorker(t *testing.T) {
	t.Helper()
	scope, err := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
	if err != nil {
		t.Fatal(err)
	}
	application := newWorkerApplication(t, api.db.AppPool, scope)

	atFixtureInstant(func() {
		ctx, cancel := context.WithCancel(t.Context())
		done := make(chan error, 1)
		go func() { done <- application.Run(ctx) }()
		defer func() {
			cancel()
			if err := <-done; err != nil {
				t.Error(err)
			}
		}()

		deadline := time.Now().Add(10 * time.Second)
		for {
			var outstanding int
			err := api.db.Pool.QueryRow(t.Context(), `select count(*) from jobs
				where state in ('pending', 'running')`).Scan(&outstanding)
			if err != nil {
				t.Fatal(err)
			}
			if outstanding == 0 {
				return
			}
			if time.Now().After(deadline) {
				t.Fatalf("worker left %d jobs outstanding", outstanding)
			}
			time.Sleep(20 * time.Millisecond)
		}
	})
}

func snapshot(t *testing.T, query string, arguments ...any) string {
	t.Helper()
	var value string
	if err := api.db.Pool.QueryRow(t.Context(), query, arguments...).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func productionLoginCookie(t *testing.T, email string) string {
	t.Helper()
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, identityFixturePassword)
	response := identityRequest(t, api.handler, "tadoku/prod", "POST", "/identity/login", body, "", nil)
	requireIdentityStatus(t, response, http.StatusOK)
	return response.Result().Cookies()[0].String()
}

func seedProductionRecoveryToken(t *testing.T, userID string) string {
	t.Helper()
	raw := base64.RawURLEncoding.EncodeToString([]byte(strings.Repeat("d", 32)))
	digest := sha256.Sum256([]byte(raw))
	_, err := api.db.Pool.Exec(t.Context(), `insert into recovery_tokens (tenant, token_hash, user_id, expires_at)
		values ('tadoku/prod', $1, $2, $3)`, hex.EncodeToString(digest[:]), userID, fixtureInstant.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
