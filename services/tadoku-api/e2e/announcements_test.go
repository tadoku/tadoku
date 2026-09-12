package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/e2e/legacy"
	"github.com/tadoku/tadoku/services/tadoku-api/features/access"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testauth"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

type fixture struct {
	db                 *testpostgres.Database
	issuer             *testauth.Issuer
	native, legacy     http.Handler
	proxied, ketoCalls atomic.Int32
	registry           *prometheus.Registry
}

func newFixture(t testing.TB) *fixture {
	t.Helper()
	f := &fixture{db: testpostgres.New(t), issuer: testauth.New(t), registry: prometheus.NewRegistry()}
	keto := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.ketoCalls.Add(1)
		if r.URL.Path != "/relation-tuples/check/openapi" || r.URL.Query().Get("namespace") != "app" || r.URL.Query().Get("object") != "tadoku" {
			t.Errorf("unexpected Keto request: %s", r.URL)
		}
		subject, relation := r.URL.Query().Get("subject_id"), r.URL.Query().Get("relation")
		if subject == "keto-unavailable" {
			w.WriteHeader(503)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		allowed := (relation == "admins" && (subject == "admin" || subject == "banned-admin")) || (relation == "banned" && (subject == "banned" || subject == "banned-admin"))
		_ = json.NewEncoder(w).Encode(map[string]bool{"allowed": allowed})
	}))
	t.Cleanup(keto.Close)
	f.legacy = legacy.ContentHandler(f.db.SQL, f.issuer.URL, keto.URL)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.proxied.Add(1)
		w.Header().Set("X-Legacy", "yes")
		f.legacy.ServeHTTP(w, r)
	}))
	t.Cleanup(upstream.Close)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	auth, err := transport.NewAuthenticator(context.Background(), f.issuer.URL, &http.Client{Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(auth.Close)
	native := app.New(content.NewService(content.NewRepository(f.db.Pool)), access.NewService(keto.URL), logger)
	f.native, err = transport.NewHandler(native, auth, f.db.Pool.Ping, transport.Upstreams{Authz: upstream.URL, Content: upstream.URL, Immersion: upstream.URL, Profile: upstream.URL}, http.DefaultTransport, 2*time.Second, f.registry, logger)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestNativeReadMatchesLegacyContract(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	now := timex.Now().Truncate(time.Second)
	for _, namespace := range []string{"main", "a/b", "a%2Fb", "hello world", "hello%20world", "日本語", "%E6%97%A5%E6%9C%AC%E8%AA%9E"} {
		id := f.db.SeedAnnouncement(t, namespace, namespace, now.Add(-time.Hour), now.Add(time.Hour), false)
		if _, err := f.db.Pool.Exec(context.Background(), "update announcements set href = $1 where id = $2", "https://tadoku.app/?x=1&y=2", id); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name, subject, identityType string
		audience                    []string
		path                        string
		status                      int
		tokenMode                   string
	}{
		{name: "guest", subject: "guest", status: 200},
		{name: "user", subject: "reader", status: 200},
		{name: "admin", subject: "admin", status: 200},
		{name: "banned", subject: "banned", status: 403},
		{name: "banned admin", subject: "banned-admin", status: 403},
		{name: "existing Keto fail-open", subject: "keto-unavailable", status: 200},
		{name: "service", subject: "system:serviceaccount:dev:caller", identityType: "service", audience: []string{"content-api"}, status: 200},
		{name: "service wrong audience", subject: "caller", identityType: "service", audience: []string{"immersion-api"}, status: 403},
		{name: "service missing audience", subject: "caller", identityType: "service", status: 403},
		{name: "multiple audiences", subject: "caller", identityType: "service", audience: []string{"immersion-api", "content-api"}, status: 200},
		{name: "user audience not enforced", subject: "reader", audience: []string{"immersion-api"}, status: 200},
		{name: "no credentials", status: 400, tokenMode: "missing"},
		{name: "malformed scheme", status: 400, tokenMode: "scheme"},
		{name: "bad signature", subject: "reader", status: 401, tokenMode: "signature"},
		{name: "expired", subject: "reader", status: 401, tokenMode: "expired"},
		{name: "future issued-at", subject: "reader", status: 401, tokenMode: "future"},
		{name: "missing user issued-at", subject: "reader", status: 500, tokenMode: "no-iat"},
		{name: "service without issued-at", subject: "caller", identityType: "service", audience: []string{"content-api"}, status: 200, tokenMode: "no-iat"},
		{name: "legacy issuer not enforced", subject: "guest", status: 200, tokenMode: "issuer"},
		{name: "lowercase bearer", subject: "guest", status: 200, tokenMode: "lowercase"},
		{name: "second authorization value", subject: "guest", status: 200, tokenMode: "second"},
		{name: "empty namespace result", subject: "guest", status: 200, path: "empty"},
		{name: "escaped slash", subject: "guest", status: 200, path: "a%2Fb"},
		{name: "escaped space", subject: "guest", status: 200, path: "hello%20world"},
		{name: "unicode", subject: "guest", status: 200, path: "%E6%97%A5%E6%9C%AC%E8%AA%9E"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims := jwt.MapClaims{"sub": tc.subject, "type": tc.identityType, "aud": tc.audience, "iat": 1700000000}
			if tc.tokenMode == "expired" {
				claims["exp"] = 1
			}
			if tc.tokenMode == "future" {
				claims["iat"] = 4102444800
			}
			if tc.tokenMode == "no-iat" {
				delete(claims, "iat")
			}
			if tc.tokenMode == "issuer" {
				claims["iss"] = "synthetic-unchecked-issuer"
			}
			token := f.issuer.Token(t, claims)
			switch tc.tokenMode {
			case "missing":
				token = ""
			case "scheme":
				token = "Basic abc"
			case "signature":
				token += "invalid"
			case "lowercase":
				token = "bearer " + strings.TrimPrefix(token, "Bearer ")
			}
			namespace := tc.path
			if namespace == "" {
				namespace = "main"
			}
			path := "/announcements/" + namespace + "/active"
			responses := make([]*httptest.ResponseRecorder, 0, 2)
			before := f.proxied.Load()
			for index, handler := range []http.Handler{f.legacy, f.native} {
				requestPath := path
				if index == 1 {
					requestPath = "/content" + path
				}
				r := httptest.NewRequest("GET", requestPath, nil)
				if tc.tokenMode == "second" {
					r.Header.Add("Authorization", "Bearer invalid")
				}
				r.Header.Add("Authorization", token)
				r.Header.Set("X-User-Id", "admin") // never trusted for identity
				r.Header.Set("X-Request-Id", "contract-test")
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, r)
				responses = append(responses, w)
			}
			direct, native := responses[0], responses[1]
			if direct.Code != tc.status {
				t.Fatalf("legacy status=%d want=%d: %s", direct.Code, tc.status, direct.Body)
			}
			if native.Code != direct.Code {
				t.Errorf("native status=%d legacy=%d", native.Code, direct.Code)
			}
			if native.Body.String() != direct.Body.String() {
				t.Errorf("native body=%s\nlegacy body=%s", native.Body, direct.Body)
			}
			if native.Header().Get("Content-Type") != direct.Header().Get("Content-Type") {
				t.Errorf("Content-Type: native=%q legacy=%q", native.Header().Get("Content-Type"), direct.Header().Get("Content-Type"))
			}
			if native.Header().Get("X-Request-Id") != "contract-test" {
				t.Error("missing correlation ID")
			}
			if f.proxied.Load() != before {
				t.Error("native request fell back to legacy")
			}
			if tc.path == "empty" && native.Body.String() != "{\"announcements\":[]}\n" {
				t.Errorf("empty representation: %s", native.Body)
			}
		})
	}
}

func TestPublicationBoundariesAndOrdering(t *testing.T) {
	// The fixed business clock is process-global: this test must stay sequential.
	f := newFixture(t)
	cutoff := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		title      string
		start, end time.Time
		deleted    bool
	}{
		{"starts exactly now", cutoff, cutoff.Add(time.Hour), false},
		{"ends exactly now", cutoff.Add(-time.Hour), cutoff, false},
		{"future", cutoff.Add(time.Microsecond), cutoff.Add(time.Hour), false},
		{"deleted", cutoff.Add(-time.Minute), cutoff.Add(time.Hour), true},
	} {
		f.db.SeedAnnouncement(t, "main", tc.title, tc.start, tc.end, tc.deleted)
	}
	for i := 1; i <= 12; i++ {
		f.db.SeedAnnouncement(t, "main", fmt.Sprintf("older %02d", i), cutoff.Add(-time.Duration(i)*time.Minute), cutoff.Add(time.Hour), false)
	}
	f.db.SeedAnnouncement(t, "other", "other namespace", cutoff, cutoff.Add(time.Hour), false)
	timex.TheWorld(cutoff, func() {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/content/announcements/main/active", nil)
		r.Header.Set("Authorization", f.issuer.Token(t, jwt.MapClaims{"sub": "guest", "iat": 1700000000}))
		f.native.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("status=%d body=%s", w.Code, w.Body)
		}
		var response openapi.ContentAnnouncements
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if len(response.Announcements) != 10 {
			t.Fatalf("got %d announcements, want 10", len(response.Announcements))
		}
		for i, item := range response.Announcements {
			want := "starts exactly now"
			if i > 0 {
				want = fmt.Sprintf("older %02d", i)
			}
			if item.Title != want {
				t.Errorf("item %d=%q want=%q", i, item.Title, want)
			}
			if item.Href != nil {
				t.Errorf("item %d href should remain null", i)
			}
		}
	})
}

func TestReadFailureDoesNotFallBackAndReadinessFails(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.db.Pool.Close()
	r := httptest.NewRequest("GET", "/content/announcements/main/active", nil)
	r.Header.Set("Authorization", f.issuer.Token(t, jwt.MapClaims{"sub": "guest", "iat": 1700000000}))
	w := httptest.NewRecorder()
	f.native.ServeHTTP(w, r)
	if w.Code != 500 || w.Body.Len() != 0 {
		t.Errorf("database failure: status=%d body=%s", w.Code, w.Body)
	}
	if f.proxied.Load() != 0 {
		t.Error("failed read fell back to legacy")
	}
	for path, status := range map[string]int{"/livez": 200, "/readyz": 503} {
		w := httptest.NewRecorder()
		f.native.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != status {
			t.Errorf("%s status=%d want=%d", path, w.Code, status)
		}
	}
	metrics, err := f.registry.Gather()
	if err != nil {
		t.Fatal(err)
	}
	var samples uint64
	for _, family := range metrics {
		if family.GetName() == "tadoku_api_native_request_duration_seconds" {
			for _, metric := range family.Metric {
				samples += metric.GetHistogram().GetSampleCount()
			}
		}
	}
	if samples != 1 {
		t.Errorf("native histogram samples=%d want=1", samples)
	}
}

func TestUnclaimedMethodsRemainProxied(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	for _, method := range []string{"HEAD", "OPTIONS", "POST", "DELETE", "PATCH"} {
		t.Run(method, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(method, "/content/announcements/main/active", nil)
			r.Header.Set("Authorization", f.issuer.Token(t, jwt.MapClaims{"sub": "guest", "iat": 1700000000}))
			f.native.ServeHTTP(w, r)
			if w.Header().Get("X-Legacy") != "yes" {
				t.Errorf("%s was not proxied", method)
			}
		})
	}
	if f.proxied.Load() != 5 {
		t.Errorf("proxy calls=%d want=5", f.proxied.Load())
	}
}

func TestCanceledNativeReadDoesNotFallBack(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest("GET", "/content/announcements/main/active", nil).WithContext(ctx)
	r.Header.Set("Authorization", f.issuer.Token(t, jwt.MapClaims{"sub": "guest", "iat": 1700000000}))
	w := httptest.NewRecorder()
	f.native.ServeHTTP(w, r)
	if w.Code != 500 || f.proxied.Load() != 0 {
		t.Errorf("canceled read: status=%d proxy calls=%d", w.Code, f.proxied.Load())
	}
}

// Run with -test.run=^$ -test.bench=BenchmarkAnnouncementRead -test.benchtime=1000x.
// Alternating paired requests use the same database/data/token and warmed real
// constructors. This is local rehearsal evidence, not production latency proof.
func BenchmarkAnnouncementRead(b *testing.B) {
	f := newFixture(b)
	now := timex.Now().Truncate(time.Second)
	for i := 0; i < 10; i++ {
		f.db.SeedAnnouncement(b, "main", fmt.Sprintf("item %d", i), now.Add(-time.Duration(i+1)*time.Hour), now.Add(time.Hour), false)
	}
	token := f.issuer.Token(b, jwt.MapClaims{"sub": "guest", "iat": 1700000000})
	requests := []*http.Request{httptest.NewRequest("GET", "/announcements/main/active", nil), httptest.NewRequest("GET", "/content/announcements/main/active", nil)}
	for _, r := range requests {
		r.Header.Set("Authorization", token)
	}
	handlers := []http.Handler{f.legacy, f.native}
	for i := 0; i < 100; i++ {
		for j, handler := range handlers {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, requests[j].Clone(context.Background()))
			if w.Code != 200 {
				b.Fatal(w.Code)
			}
		}
	}
	samples := [][]time.Duration{make([]time.Duration, 0, b.N), make([]time.Duration, 0, b.N)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for n := 0; n < 2; n++ {
			j := (i + n) % 2
			w := httptest.NewRecorder()
			r := requests[j].Clone(context.Background())
			start := time.Now() // measurement, not business time
			handlers[j].ServeHTTP(w, r)
			samples[j] = append(samples[j], time.Since(start))
			if w.Code != 200 {
				b.Fatalf("status=%d", w.Code)
			}
		}
	}
	b.StopTimer()
	for j, name := range []string{"legacy", "native"} {
		slices.Sort(samples[j])
		b.ReportMetric(float64(samples[j][len(samples[j])/2].Nanoseconds()), name+"-p50-ns")
		b.ReportMetric(float64(samples[j][(len(samples[j])-1)*95/100].Nanoseconds()), name+"-p95-ns")
	}
}
