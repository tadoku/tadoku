package keto

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				assert.Equal(t, "App", r.URL.Query().Get("namespace"))
				assert.Equal(t, "global", r.URL.Query().Get("object"))
				assert.Equal(t, "admins", r.URL.Query().Get("relation"))
				assert.Equal(t, "User", r.URL.Query().Get("subject_set.namespace"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, server.URL)
			result, err := client.CheckPermission(context.Background(), "App", "global", "admins", Subject{Namespace: "User", Object: "test-user-id"})

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestAddRelation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/admin/relation-tuples", r.URL.Path)

		var body map[string]any
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "App", body["namespace"])
		assert.Equal(t, "global", body["object"])
		assert.Equal(t, "admins", body["relation"])

		subjectSet := body["subject_set"].(map[string]any)
		assert.Equal(t, "User", subjectSet["namespace"])
		assert.Equal(t, "test-user-id", subjectSet["object"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(body)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL)
	err := client.AddRelation(context.Background(), "App", "global", "admins", Subject{Namespace: "User", Object: "test-user-id"})

	require.NoError(t, err)
}

func TestDeleteRelation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/admin/relation-tuples", r.URL.Path)
		assert.Equal(t, "App", r.URL.Query().Get("namespace"))
		assert.Equal(t, "global", r.URL.Query().Get("object"))
		assert.Equal(t, "admins", r.URL.Query().Get("relation"))
		assert.Equal(t, "User", r.URL.Query().Get("subject_set.namespace"))
		assert.Equal(t, "test-user-id", r.URL.Query().Get("subject_set.object"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, server.URL)
	err := client.DeleteRelation(context.Background(), "App", "global", "admins", Subject{Namespace: "User", Object: "test-user-id"})

	require.NoError(t, err)
}

func TestCheckPermissions(t *testing.T) {
	t.Run("checks multiple permissions in parallel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			relation := r.URL.Query().Get("relation")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Return allowed=true for admins, allowed=false for banned
			allowed := relation == "admins"
			json.NewEncoder(w).Encode(map[string]any{"allowed": allowed})
		}))
		defer server.Close()

		client := NewClient(server.URL, server.URL)
		checks := []PermissionCheck{
			{Namespace: "App", Object: "global", Relation: "admins", Subject: Subject{Namespace: "User", Object: "user-1"}},
			{Namespace: "App", Object: "global", Relation: "banned", Subject: Subject{Namespace: "User", Object: "user-1"}},
		}

		results := client.CheckPermissions(context.Background(), checks)

		require.Len(t, results, 2)

		// Results should be in the same order as input
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
				// Return an error for the banned check
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{"error": "internal error"})
			}
		}))
		defer server.Close()

		client := NewClient(server.URL, server.URL)
		checks := []PermissionCheck{
			{Namespace: "App", Object: "global", Relation: "admins", Subject: Subject{Namespace: "User", Object: "user-1"}},
			{Namespace: "App", Object: "global", Relation: "banned", Subject: Subject{Namespace: "User", Object: "user-1"}},
		}

		results := client.CheckPermissions(context.Background(), checks)

		require.Len(t, results, 2)

		// First check should succeed
		assert.True(t, results[0].Allowed)
		require.NoError(t, results[0].Err)

		// Second check should have an error
		assert.False(t, results[1].Allowed)
		require.Error(t, results[1].Err)
	})
}
