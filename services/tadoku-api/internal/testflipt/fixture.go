package testflipt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"

	"github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type namespaceState struct {
	revision  uint64
	resources map[string]json.RawMessage
}

type Fixture struct {
	mu          sync.Mutex
	targets     flipt.Targets
	namespaces  map[flipt.Target]*namespaceState
	unavailable bool
	requests    int
	creates     map[flipt.Target]CreateCounts
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
			revision:  1,
			resources: seededResources([]string{"11111111-1111-4111-8111-111111111111"}),
		},
		{Environment: "test", Namespace: "e2e_flipt-0123abcd"}: {
			revision:  1,
			resources: seededResources([]string{}),
		},
	}
	f.unavailable = false
	f.requests = 0
	f.creates = make(map[flipt.Target]CreateCounts)
}

func (f *Fixture) SeedTenant(key tenant.Key) error {
	target, err := f.targets.Resolve(tenant.WithKey(context.Background(), key))
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.namespaces[target] == nil {
		f.namespaces[target] = &namespaceState{revision: 1, resources: seededResources([]string{})}
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
	if state == nil || request.FlagKey != "release-webv3" {
		return featureflags.ProviderResult{}, featureflags.ErrFlagNotFound
	}
	if request.Context["authenticated"] != "true" {
		return featureflags.ProviderResult{}, featureflags.ErrInvalidResponse
	}
	var segment struct {
		Constraints []struct {
			Value string `json:"value"`
		} `json:"constraints"`
	}
	err = json.Unmarshal(state.resources["flipt.core.Segment/release-webv3-access"], &segment)
	if err != nil || len(segment.Constraints) != 1 {
		return featureflags.ProviderResult{}, featureflags.ErrInvalidResponse
	}
	var members []string
	if err := json.Unmarshal([]byte(segment.Constraints[0].Value), &members); err != nil {
		return featureflags.ProviderResult{}, featureflags.ErrInvalidResponse
	}
	enabled := slices.Contains(members, request.EntityID)
	return featureflags.ProviderResult{Enabled: enabled, Reason: "MATCH_EVALUATION_REASON"}, nil
}

func (f *Fixture) serveHTTP(response http.ResponseWriter, request *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests++
	if f.unavailable {
		http.Error(response, "unavailable", http.StatusServiceUnavailable)
		return
	}

	parts := strings.Split(strings.TrimPrefix(request.URL.Path, "/"), "/")
	if f.serveLifecycle(response, request, parts) {
		return
	}
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

	if request.Method == http.MethodPut && len(parts) == 7 {
		var update struct {
			Key      string          `json:"key"`
			Revision string          `json:"revision"`
			Payload  json.RawMessage `json:"payload"`
		}
		if err := json.NewDecoder(request.Body).Decode(&update); err != nil || !json.Valid(update.Payload) {
			http.Error(response, "invalid update", http.StatusBadRequest)
			return
		}
		identity := "flipt.core.Segment/" + update.Key
		if _, exists := state.resources[identity]; !exists {
			http.NotFound(response, request)
			return
		}
		if update.Revision != revisionString(state) {
			response.WriteHeader(http.StatusConflict)
			return
		}
		state.resources[identity] = update.Payload
		state.revision++
		f.writeResource(response, target, state, update.Key, update.Payload)
		return
	}
	http.NotFound(response, request)
}

func (f *Fixture) writeResource(
	response http.ResponseWriter,
	target flipt.Target,
	state *namespaceState,
	key string,
	payload json.RawMessage,
) {
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(map[string]any{
		"resource": map[string]any{
			"namespaceKey": target.Namespace,
			"key":          key,
			"payload":      payload,
		},
		"revision": revisionString(state),
	})
}

func revisionString(state *namespaceState) string { return fmt.Sprintf("%040x", state.revision) }
