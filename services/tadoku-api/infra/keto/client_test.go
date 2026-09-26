package keto

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReadClientUsesHTTPClientTimeout(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	t.Cleanup(server.CloseClientConnections)

	client := NewReadClient(server.URL, WithHTTPClient(&http.Client{Timeout: 50 * time.Millisecond}))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	requestStarted := time.Now()
	_, err := client.CheckPermission(
		ctx,
		"app",
		"tadoku",
		"banned",
		Subject{ID: "user"},
	)

	select {
	case <-started:
	default:
		t.Fatal("Keto provider did not receive the request")
	}
	if err == nil {
		t.Fatal("permission check unexpectedly succeeded")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("permission check error=%v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(requestStarted); elapsed > time.Second {
		t.Fatalf("permission check took %v, want less than 1s", elapsed)
	}
}

func TestNewReadClientUsesHTTPClientConnectionPool(t *testing.T) {
	const concurrentChecks = 5

	started := [2]chan struct{}{make(chan struct{}, concurrentChecks), make(chan struct{}, concurrentChecks)}
	release := [2]chan struct{}{make(chan struct{}), make(chan struct{})}
	var connections atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wave := 0
		if r.URL.Query().Get("relation") == "second" {
			wave = 1
		}
		started[wave] <- struct{}{}
		<-release[wave]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"allowed":true}`))
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			connections.Add(1)
		}
	}
	server.Start()
	t.Cleanup(server.Close)
	t.Cleanup(func() {
		for _, ch := range release {
			select {
			case <-ch:
			default:
				close(ch)
			}
		}
	})

	transport := &http.Transport{
		MaxIdleConns:        concurrentChecks,
		MaxIdleConnsPerHost: concurrentChecks,
	}
	t.Cleanup(transport.CloseIdleConnections)
	client := NewReadClient(server.URL, WithHTTPClient(&http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}))

	runWave := func(relation string, wave int) {
		t.Helper()
		errs := make(chan error, concurrentChecks)
		for range concurrentChecks {
			go func() {
				_, err := client.CheckPermission(
					context.Background(),
					"app",
					"tadoku",
					relation,
					Subject{ID: "user"},
				)
				errs <- err
			}()
		}

		for range concurrentChecks {
			select {
			case <-started[wave]:
			case <-time.After(5 * time.Second):
				t.Fatal("Keto provider did not receive all concurrent requests")
			}
		}
		close(release[wave])
		for range concurrentChecks {
			if err := <-errs; err != nil {
				t.Errorf("permission check: %v", err)
			}
		}
	}

	runWave("first", 0)
	runWave("second", 1)

	if got := connections.Load(); got != concurrentChecks {
		t.Errorf("new connections=%d, want %d", got, concurrentChecks)
	}
}

func TestCheckPermissionPropagatesCancellation(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		close(started)
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	t.Cleanup(server.CloseClientConnections)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, err := NewReadClient(server.URL).CheckPermission(
			ctx,
			"app",
			"tadoku",
			"banned",
			Subject{ID: "user"},
		)
		result <- err
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("Keto client did not start the HTTP request")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("permission check error=%v, want context canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Keto client did not return after cancellation")
	}
}

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   map[string]any
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "permission allowed",
			statusCode:     http.StatusOK,
			responseBody:   map[string]any{"allowed": true},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:           "permission denied",
			statusCode:     http.StatusOK,
			responseBody:   map[string]any{"allowed": false},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "forbidden returns false",
			statusCode:     http.StatusForbidden,
			responseBody:   map[string]any{"allowed": false},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:           "internal server error returns error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   map[string]any{"error": "internal server error"},
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/relation-tuples/check/openapi", r.URL.Path)
				assert.Equal(t, "app", r.URL.Query().Get("namespace"))
				assert.Equal(t, "global", r.URL.Query().Get("object"))
				assert.Equal(t, "admins", r.URL.Query().Get("relation"))
				assert.Equal(t, "test-user-id", r.URL.Query().Get("subject_id"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, server.URL)
			result, err := client.CheckPermission(context.Background(), "app", "global", "admins", Subject{ID: "test-user-id"})

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestCheckPermission_subjectSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/relation-tuples/check/openapi", r.URL.Path)
		assert.Equal(t, "app", r.URL.Query().Get("namespace"))
		assert.Equal(t, "global", r.URL.Query().Get("object"))
		assert.Equal(t, "admins", r.URL.Query().Get("relation"))

		assert.Empty(t, r.URL.Query().Get("subject_id"))
		assert.Equal(t, "Group", r.URL.Query().Get("subject_set.namespace"))
		assert.Equal(t, "admins", r.URL.Query().Get("subject_set.object"))
		assert.Equal(t, "member", r.URL.Query().Get("subject_set.relation"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"allowed": true})
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL)
	allowed, err := client.CheckPermission(
		context.Background(),
		"app",
		"global",
		"admins",
		Subject{Set: &SubjectSet{Namespace: "Group", Object: "admins", Relation: "member"}},
	)

	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCheckPermission_invalidSubject(t *testing.T) {
	client := NewClient("http://localhost", "http://localhost")
	allowed, err := client.CheckPermission(context.Background(), "app", "global", "admins", Subject{})
	require.Error(t, err)
	assert.False(t, allowed)
}

func TestAddRelation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/admin/relation-tuples", r.URL.Path)

		var body map[string]any
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "app", body["namespace"])
		assert.Equal(t, "global", body["object"])
		assert.Equal(t, "admins", body["relation"])

		assert.Equal(t, "test-user-id", body["subject_id"])
		_, ok := body["subject_set"]
		require.False(t, ok, "subject_set should be omitted when subject_id is set")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL)
	err := client.AddRelation(context.Background(), "app", "global", "admins", Subject{ID: "test-user-id"})

	require.NoError(t, err)
}

func TestAddRelation_subjectSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/admin/relation-tuples", r.URL.Path)

		var body map[string]any
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "app", body["namespace"])
		assert.Equal(t, "global", body["object"])
		assert.Equal(t, "admins", body["relation"])

		assert.Empty(t, body["subject_id"])
		subjectSet, ok := body["subject_set"].(map[string]any)
		require.True(t, ok, "subject_set should be a map")
		assert.Equal(t, "Group", subjectSet["namespace"])
		assert.Equal(t, "admins", subjectSet["object"])
		assert.Equal(t, "member", subjectSet["relation"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL)
	err := client.AddRelation(
		context.Background(),
		"app",
		"global",
		"admins",
		Subject{Set: &SubjectSet{Namespace: "Group", Object: "admins", Relation: "member"}},
	)

	require.NoError(t, err)
}

func TestAddRelation_invalidSubject(t *testing.T) {
	client := NewClient("http://localhost", "http://localhost")
	err := client.AddRelation(context.Background(), "app", "global", "admins", Subject{})
	require.Error(t, err)
}

func TestDeleteRelation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/admin/relation-tuples", r.URL.Path)
		assert.Equal(t, "app", r.URL.Query().Get("namespace"))
		assert.Equal(t, "global", r.URL.Query().Get("object"))
		assert.Equal(t, "admins", r.URL.Query().Get("relation"))
		assert.Equal(t, "test-user-id", r.URL.Query().Get("subject_id"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL)
	err := client.DeleteRelation(context.Background(), "app", "global", "admins", Subject{ID: "test-user-id"})

	require.NoError(t, err)
}

func TestDeleteRelation_subjectSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/admin/relation-tuples", r.URL.Path)
		assert.Equal(t, "app", r.URL.Query().Get("namespace"))
		assert.Equal(t, "global", r.URL.Query().Get("object"))
		assert.Equal(t, "admins", r.URL.Query().Get("relation"))

		assert.Empty(t, r.URL.Query().Get("subject_id"))
		assert.Equal(t, "Group", r.URL.Query().Get("subject_set.namespace"))
		assert.Equal(t, "admins", r.URL.Query().Get("subject_set.object"))
		assert.Equal(t, "member", r.URL.Query().Get("subject_set.relation"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL)
	err := client.DeleteRelation(
		context.Background(),
		"app",
		"global",
		"admins",
		Subject{Set: &SubjectSet{Namespace: "Group", Object: "admins", Relation: "member"}},
	)

	require.NoError(t, err)
}

func TestDeleteRelation_invalidSubject(t *testing.T) {
	client := NewClient("http://localhost", "http://localhost")
	err := client.DeleteRelation(context.Background(), "app", "global", "admins", Subject{})
	require.Error(t, err)
}

func TestCheckPermissions(t *testing.T) {
	t.Run("checks multiple permissions in parallel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			relation := r.URL.Query().Get("relation")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			allowed := relation == "admins"
			json.NewEncoder(w).Encode(map[string]any{"allowed": allowed})
		}))
		defer server.Close()

		client := NewClient(server.URL, server.URL)
		checks := []PermissionCheck{
			{Namespace: "app", Object: "global", Relation: "admins", Subject: Subject{ID: "user-1"}},
			{Namespace: "app", Object: "global", Relation: "banned", Subject: Subject{ID: "user-1"}},
		}

		results := client.CheckPermissions(context.Background(), checks)

		require.Len(t, results, 2)

		assert.Equal(t, "admins", results[0].Check.Relation)
		assert.True(t, results[0].Allowed)
		require.NoError(t, results[0].Err)

		assert.Equal(t, "banned", results[1].Check.Relation)
		assert.False(t, results[1].Allowed)
		require.NoError(t, results[1].Err)
	})

	t.Run("handles empty checks slice", func(t *testing.T) {
		client := NewClient("http://localhost", "http://localhost")
		results := client.CheckPermissions(context.Background(), []PermissionCheck{})

		assert.Empty(t, results)
	})

	t.Run("handles errors for individual checks", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			relation := r.URL.Query().Get("relation")

			w.Header().Set("Content-Type", "application/json")

			if relation == "admins" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]any{"allowed": true})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{"error": "internal error"})
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, server.URL)
		checks := []PermissionCheck{
			{Namespace: "app", Object: "global", Relation: "admins", Subject: Subject{ID: "user-1"}},
			{Namespace: "app", Object: "global", Relation: "banned", Subject: Subject{ID: "user-1"}},
		}

		results := client.CheckPermissions(context.Background(), checks)

		require.Len(t, results, 2)

		assert.True(t, results[0].Allowed)
		require.NoError(t, results[0].Err)

		assert.False(t, results[1].Allowed)
		require.Error(t, results[1].Err)
	})
}
