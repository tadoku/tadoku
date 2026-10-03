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

	"github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type namespaceState struct {
	members  map[string]struct{}
	revision uint64
}

type Fixture struct {
	mu          sync.Mutex
	targets     flipt.Targets
	namespaces  map[flipt.Target]*namespaceState
	unavailable bool
	server      *httptest.Server
}

func New() *Fixture {
	targets, _ := flipt.NewTargets("local", "default", "test")
	f := &Fixture{targets: targets}
	f.server = httptest.NewServer(http.HandlerFunc(f.serveHTTP))
	f.Reset()
	return f
}

func (f *Fixture) Close() { f.server.Close() }

func (f *Fixture) URL() string { return f.server.URL }

func (f *Fixture) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.namespaces = map[flipt.Target]*namespaceState{
		{Environment: "local", Namespace: "default"}: {
			members:  map[string]struct{}{"11111111-1111-4111-8111-111111111111": {}},
			revision: 1,
		},
		{Environment: "test", Namespace: "e2e_flipt-0123abcd"}: {
			members:  make(map[string]struct{}),
			revision: 1,
		},
	}
	f.unavailable = false
}

func (f *Fixture) SeedTenant(key tenant.Key) error {
	target, err := f.targets.Resolve(tenant.WithKey(context.Background(), key))
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.namespaces[target] == nil {
		f.namespaces[target] = &namespaceState{members: make(map[string]struct{}), revision: 1}
	}
	return nil
}

func (f *Fixture) SetUnavailable(unavailable bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unavailable = unavailable
}

func (f *Fixture) EvaluateBoolean(
	ctx context.Context,
	request featureflags.EvaluationRequest,
) (featureflags.ProviderResult, error) {
	if err := ctx.Err(); err != nil {
		return featureflags.ProviderResult{}, err
	}
	target, err := f.targets.Resolve(ctx)
	if err != nil {
		return featureflags.ProviderResult{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unavailable {
		return featureflags.ProviderResult{}, fmt.Errorf("test Flipt unavailable")
	}
	state := f.namespaces[target]
	if state == nil || request.FlagKey != "release-log-entry-v2" {
		return featureflags.ProviderResult{}, featureflags.ErrFlagNotFound
	}
	if request.Context["authenticated"] != "true" {
		return featureflags.ProviderResult{}, featureflags.ErrInvalidResponse
	}
	_, enabled := state.members[request.EntityID]
	return featureflags.ProviderResult{Enabled: enabled, Reason: "MATCH_EVALUATION_REASON"}, nil
}

func (f *Fixture) serveHTTP(response http.ResponseWriter, request *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unavailable {
		http.Error(response, "unavailable", http.StatusServiceUnavailable)
		return
	}

	parts := strings.Split(strings.TrimPrefix(request.URL.Path, "/"), "/")
	if len(parts) < 7 || parts[0] != "api" || parts[1] != "v2" || parts[2] != "environments" ||
		parts[4] != "namespaces" || parts[6] != "resources" {
		http.NotFound(response, request)
		return
	}
	target := flipt.Target{Environment: parts[3], Namespace: parts[5]}
	state := f.namespaces[target]
	if state == nil {
		http.NotFound(response, request)
		return
	}

	switch {
	case request.Method == http.MethodGet && len(parts) == 9 && parts[7] == "flipt.core.Segment" &&
		parts[8] == "release-log-entry-v2-access":
		f.writeSegment(response, target, state)
	case request.Method == http.MethodPut && len(parts) == 7:
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
		if update.Revision != revisionString(state) {
			response.WriteHeader(http.StatusConflict)
			return
		}
		var members []string
		if err := json.Unmarshal([]byte(update.Payload.Constraints[0].Value), &members); err != nil {
			http.Error(response, "invalid members", http.StatusBadRequest)
			return
		}
		state.members = make(map[string]struct{}, len(members))
		for _, member := range members {
			state.members[member] = struct{}{}
		}
		state.revision++
		f.writeSegment(response, target, state)
	default:
		http.NotFound(response, request)
	}
}

func (f *Fixture) writeSegment(response http.ResponseWriter, target flipt.Target, state *namespaceState) {
	members := make([]string, 0, len(state.members))
	for member := range state.members {
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
			"type":        "ENTITY_ID_COMPARISON_TYPE",
			"property":    "entityId",
			"operator":    "isoneof",
			"value":       string(encodedMembers),
			"description": "",
		}},
	}
	encodedPayload, _ := json.Marshal(payload)
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]any{
		"resource": map[string]any{
			"namespaceKey": target.Namespace,
			"key":          "release-log-entry-v2-access",
			"payload":      json.RawMessage(encodedPayload),
		},
		"revision": revisionString(state),
	})
}

func revisionString(state *namespaceState) string { return fmt.Sprintf("%040x", state.revision) }
