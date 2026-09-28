package testflipt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
)

type Fixture struct {
	mu          sync.Mutex
	members     map[string]struct{}
	revision    uint64
	unavailable bool
	server      *httptest.Server
}

func New() *Fixture {
	f := &Fixture{}
	f.server = httptest.NewServer(http.HandlerFunc(f.serveHTTP))
	f.Reset()
	return f
}

func (f *Fixture) Close() { f.server.Close() }

func (f *Fixture) URL() string { return f.server.URL }

func (f *Fixture) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.members = map[string]struct{}{
		"11111111-1111-4111-8111-111111111111": {},
	}
	f.revision = 1
	f.unavailable = false
}

func (f *Fixture) SetUnavailable(unavailable bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unavailable = unavailable
}

func (f *Fixture) EvaluateBoolean(ctx context.Context, request featureflags.EvaluationRequest) (featureflags.ProviderResult, error) {
	if err := ctx.Err(); err != nil {
		return featureflags.ProviderResult{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unavailable {
		return featureflags.ProviderResult{}, fmt.Errorf("test Flipt unavailable")
	}
	if request.FlagKey != "release-log-entry-v2" {
		return featureflags.ProviderResult{}, featureflags.ErrFlagNotFound
	}
	if request.Context["authenticated"] != "true" {
		return featureflags.ProviderResult{}, featureflags.ErrInvalidResponse
	}
	_, enabled := f.members[request.EntityID]
	return featureflags.ProviderResult{Enabled: enabled, Reason: "MATCH_EVALUATION_REASON"}, nil
}

func (f *Fixture) serveHTTP(response http.ResponseWriter, request *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unavailable {
		http.Error(response, "unavailable", http.StatusServiceUnavailable)
		return
	}

	switch {
	case request.Method == http.MethodGet && strings.HasSuffix(request.URL.Path, "/resources/flipt.core.Segment/release-log-entry-v2-access"):
		f.writeSegment(response)
	case request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/resources"):
		var update struct {
			Revision string `json:"revision"`
			Payload  struct {
				Constraints []struct {
					Value string `json:"value"`
				} `json:"constraints"`
			} `json:"payload"`
		}
		if err := json.NewDecoder(request.Body).Decode(&update); err != nil || len(update.Payload.Constraints) != 1 {
			http.Error(response, "invalid update", http.StatusBadRequest)
			return
		}
		if update.Revision != f.revisionString() {
			response.WriteHeader(http.StatusConflict)
			return
		}
		var members []string
		if err := json.Unmarshal([]byte(update.Payload.Constraints[0].Value), &members); err != nil {
			http.Error(response, "invalid members", http.StatusBadRequest)
			return
		}
		f.members = make(map[string]struct{}, len(members))
		for _, member := range members {
			f.members[member] = struct{}{}
		}
		f.revision++
		f.writeSegment(response)
	default:
		http.NotFound(response, request)
	}
}

func (f *Fixture) writeSegment(response http.ResponseWriter) {
	members := make([]string, 0, len(f.members))
	for member := range f.members {
		members = append(members, member)
	}
	sort.Strings(members)
	encodedMembers, _ := json.Marshal(members)
	payload := map[string]any{
		"@type":       "flipt.core.Segment",
		"key":         "release-log-entry-v2-access",
		"name":        "Release log entry v2 access",
		"description": "Kratos UUIDs explicitly granted access to release log entry v2.",
		"matchType":   "ALL_MATCH_TYPE",
		"constraints": []map[string]string{{
			"type": "ENTITY_ID_COMPARISON_TYPE", "property": "entityId", "operator": "isoneof", "value": string(encodedMembers), "description": "",
		}},
	}
	encodedPayload, _ := json.Marshal(payload)
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]any{
		"resource": map[string]any{"namespaceKey": "default", "key": "release-log-entry-v2-access", "payload": json.RawMessage(encodedPayload)},
		"revision": f.revisionString(),
	})
}

func (f *Fixture) revisionString() string { return fmt.Sprintf("%040x", f.revision) }
