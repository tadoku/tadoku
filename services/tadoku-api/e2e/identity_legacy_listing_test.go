package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	accountidentity "github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
)

func TestIdentityLegacyAdminSingleWalk(t *testing.T) {
	for _, size := range []int{499, 1200, 4527} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			api.reset(t, filepath.Join("testdata", "ProfileUsersList", "200_admin"))
			t.Cleanup(api.resetProfileCaches)
			var calls atomic.Int64
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != http.MethodGet || r.URL.Path != "/admin/identities" {
					t.Errorf("unexpected provider request: %s %s", r.Method, r.URL)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				limit, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
				offset, _ := strconv.Atoi(r.URL.Query().Get("page_token"))
				if limit != 500 {
					t.Errorf("provider page size=%d", limit)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				end := min(offset+limit, size)
				rows := make([]map[string]any, 0, end-offset)
				for i := offset; i < end; i++ {
					rows = append(rows, map[string]any{
						"id":         fmt.Sprintf("aaaaaaaa-aaaa-4aaa-8aaa-%012d", size-i),
						"schema_id":  "user",
						"created_at": "2020-01-01T00:00:00Z",
						"traits": map[string]string{
							"display_name": fmt.Sprintf("Provider %d", i),
							"email":        fmt.Sprintf("person%04d@example.test", i),
						},
					})
				}
				if end < size {
					w.Header().Set("Link", fmt.Sprintf("</admin/identities?page_token=%d>; rel=\"next\"", end))
				}
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(rows); err != nil {
					t.Error(err)
				}
			}))
			defer provider.Close()
			legacy := accountidentity.NewLegacyReader(
				provider.URL,
				accountidentity.WithHTTPClient(&http.Client{Timeout: 2 * time.Second}),
			)
			reader := accountidentity.NewReader(api.db.AppPool, legacy, featureflags.NewEvaluator(api.flipt, nil))
			*api.profile = *profile.NewService(profile.NewRepository(api.db.AppPool), api.roles, reader)
			admin := readHTTPRequest(t, filepath.Join("testdata", "ProfileUsersList", "200_admin"))

			for _, scenario := range []struct {
				path, first  string
				total, count int
			}{
				{"/profile/users?page_size=20&page=1", "Provider 20", size, 20},
				{"/profile/users?query=person0123@example.test", "Provider 123", 1, 1},
			} {
				calls.Store(0)
				request := httptest.NewRequest(http.MethodGet, scenario.path, nil)
				request.Header = admin.Header.Clone()
				response := httptest.NewRecorder()
				atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })
				var result struct {
					Users []struct {
						Name string `json:"display_name"`
					} `json:"users"`
					Total int `json:"total_size"`
				}
				if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &result) != nil {
					t.Fatalf("list=%d %s", response.Code, response.Body)
				}
				if result.Total != scenario.total || len(result.Users) != scenario.count ||
					result.Users[0].Name != scenario.first {
					t.Fatalf("list result=%+v", result)
				}

				wantCalls := (size + 499) / 500
				t.Logf("identities=%d request=%s provider_calls=%d single_walk=%d", size, scenario.path, calls.Load(), wantCalls)
				if calls.Load() != int64(wantCalls) {
					t.Errorf("provider calls=%d, want one complete walk=%d", calls.Load(), wantCalls)
				}
			}
		})
	}
}
