package e2e_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
)

func TestIdentityV2UniqueAccountTenantIsolation(t *testing.T) {
	resetIdentity(t)
	registration := `{"email":"unique@example.test","display_name":"Unique Reader","password":"Independent-653!"}`
	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/registration", registration, "", nil)
	requireIdentityStatus(t, response, http.StatusCreated)
	for _, key := range []string{"tadoku/prod", identityBranchB} {
		response = identityRequest(t, api.handler, key, "POST", "/identity/login",
			`{"email":"unique@example.test","password":"Independent-653!"}`, "", nil)
		requireIdentityStatus(t, response, http.StatusUnauthorized)
		var failure struct{ Code string }
		if err := json.Unmarshal(response.Body.Bytes(), &failure); err != nil {
			t.Fatal(err)
		}
		if failure.Code != "wrong_password" || response.Header().Get("Set-Cookie") != "" {
			t.Errorf("foreign tenant %s accepted the account or disclosed its existence", key)
		}
	}

	response = identityRequest(t, api.handler, identityBranchB, "POST", "/identity/registration", registration, "", nil)
	requireIdentityStatus(t, response, http.StatusCreated)
}

func TestIdentityV2DuplicateRegistrationIsAtomic(t *testing.T) {
	resetIdentity(t)
	var before, after int
	err := api.db.Pool.QueryRow(t.Context(), "select count(*) from users where tenant = $1", identityBranchA).Scan(&before)
	if err != nil {
		t.Fatal(err)
	}

	response := identityRequest(t, api.handler, identityBranchA, "POST", "/identity/registration",
		`{"email":"READER@EXAMPLE.TEST","display_name":"Duplicate Reader","password":"Independent-653!"}`, "", nil)
	requireIdentityStatus(t, response, http.StatusConflict)
	err = api.db.Pool.QueryRow(t.Context(), "select count(*) from users where tenant = $1", identityBranchA).Scan(&after)
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("duplicate registration left a user behind: before=%d after=%d", before, after)
	}
}

type waitingIdentityObserver struct {
	entered chan struct{}
	release chan struct{}
	waiting atomic.Bool
}

func (o *waitingIdentityObserver) ObserveEvaluation(featureflags.Observation) {
	if o.waiting.CompareAndSwap(false, true) {
		close(o.entered)
		<-o.release
	}
}

func TestIdentityRegistrationLeavesPoolAvailableDuringEvaluation(t *testing.T) {
	resetIdentity(t)
	config := api.db.AppPool.Config()
	config.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	observer := &waitingIdentityObserver{entered: make(chan struct{}), release: make(chan struct{})}
	handler, _, _, err := newTestRouterWithLogger(
		t.Context(), pool, pool, api.keto, api.kratos,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func(endpoints *routerEndpoints) { endpoints.identityObserver = observer },
	)
	if err != nil {
		t.Fatal(err)
	}

	finished := make(chan struct{})
	go func() {
		defer close(finished)
		response := identityRequest(t, handler, identityBranchA, "POST", "/identity/registration",
			`{"email":"pool@example.test","display_name":"Pool Reader","password":"Independent-653!"}`, "", nil)
		if response.Code != http.StatusCreated {
			t.Errorf("registration status=%d body=%s", response.Code, response.Body)
		}
	}()
	defer func() {
		close(observer.release)
		<-finished
	}()

	select {
	case <-observer.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("registration did not reach the flag evaluation boundary")
	}
	if held := pool.Stat().AcquiredConns(); held != 0 {
		t.Errorf("registration holds %d database connections while flag evaluation waits, want 0", held)
	} else {
		request := readHTTPRequest(t, "testdata/BannedUsers/200_guest")
		request.URL.Path = "/immersion/languages"
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		requireIdentityStatus(t, response, http.StatusOK)
	}
}
