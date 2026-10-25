package testflipt

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type CreateCounts struct {
	Namespace int
	Resources int
}

func (f *Fixture) Exists(key tenant.Key) bool {
	target, _ := f.targets.Resolve(tenant.WithKey(context.Background(), key))
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.namespaces[target] != nil
}

func (f *Fixture) ResourceCount(key tenant.Key) int {
	target, _ := f.targets.Resolve(tenant.WithKey(context.Background(), key))
	f.mu.Lock()
	defer f.mu.Unlock()
	if state := f.namespaces[target]; state != nil {
		return len(state.resources)
	}
	return 0
}

func (f *Fixture) CreateCounts(key tenant.Key) CreateCounts {
	target, _ := f.targets.Resolve(tenant.WithKey(context.Background(), key))
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.creates[target]
}

func (f *Fixture) RequestCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.requests
}

func seededResources() map[string]json.RawMessage {
	return map[string]json.RawMessage{
		"flipt.core.Segment/release-log-entry-v2-access": json.RawMessage(`{}`),
		"flipt.core.Flag/release-log-entry-v2":           json.RawMessage(`{}`),
	}
}

func (f *Fixture) serveLifecycle(response http.ResponseWriter, request *http.Request, parts []string) bool {
	if len(parts) < 5 || parts[0] != "api" || parts[1] != "v2" ||
		parts[2] != "environments" || parts[4] != "namespaces" {
		return false
	}
	response.Header().Set("Content-Type", "application/json")
	if len(parts) == 5 && request.Method == http.MethodPost {
		var namespace struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(request.Body).Decode(&namespace); err != nil || namespace.Key == "" {
			http.Error(response, "invalid namespace", http.StatusBadRequest)
			return true
		}
		target := flipt.Target{Environment: parts[3], Namespace: namespace.Key}
		if f.namespaces[target] != nil {
			response.WriteHeader(http.StatusConflict)
			return true
		}
		f.namespaces[target] = &namespaceState{
			members:   make(map[string]struct{}),
			revision:  1,
			resources: make(map[string]json.RawMessage),
		}
		counts := f.creates[target]
		counts.Namespace++
		f.creates[target] = counts
		_ = json.NewEncoder(response).Encode(map[string]any{"namespace": namespace})
		return true
	}
	if len(parts) < 6 {
		return false
	}
	target := flipt.Target{Environment: parts[3], Namespace: parts[5]}
	state := f.namespaces[target]
	if state == nil {
		if len(parts) == 6 && request.Method == http.MethodDelete {
			_, _ = response.Write([]byte(`{}`))
			return true
		}
		http.NotFound(response, request)
		return true
	}
	if len(parts) == 6 {
		switch request.Method {
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"namespace": map[string]string{"key": target.Namespace}})
		case http.MethodDelete:
			delete(f.namespaces, target)
			_, _ = response.Write([]byte(`{}`))
		default:
			http.NotFound(response, request)
		}
		return true
	}
	if parts[6] != "resources" {
		return false
	}
	if len(parts) == 9 && request.Method == http.MethodGet {
		payload, ok := state.resources[parts[7]+"/"+parts[8]]
		if !ok {
			http.NotFound(response, request)
			return true
		}
		if parts[7] == "flipt.core.Segment" && parts[8] == "release-log-entry-v2-access" {
			return false
		}
		_ = json.NewEncoder(response).Encode(map[string]any{
			"resource": map[string]any{"key": parts[8], "namespaceKey": target.Namespace, "payload": payload},
		})
		return true
	}
	if len(parts) != 7 || request.Method != http.MethodPost {
		return false
	}

	var resource struct {
		Key     string          `json:"key"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.NewDecoder(request.Body).Decode(&resource); err != nil {
		http.Error(response, "invalid resource", http.StatusBadRequest)
		return true
	}
	var payload struct {
		Type        string `json:"@type"`
		Constraints []struct {
			Value string `json:"value"`
		} `json:"constraints"`
	}
	if err := json.Unmarshal(resource.Payload, &payload); err != nil || resource.Key == "" || payload.Type == "" {
		http.Error(response, "invalid payload", http.StatusBadRequest)
		return true
	}
	identity := payload.Type + "/" + resource.Key
	if _, exists := state.resources[identity]; exists {
		response.WriteHeader(http.StatusConflict)
		return true
	}
	if payload.Type == "flipt.core.Segment" && resource.Key == "release-log-entry-v2-access" {
		if len(payload.Constraints) != 1 {
			http.Error(response, "invalid members constraint", http.StatusBadRequest)
			return true
		}
		var members []string
		if err := json.Unmarshal([]byte(payload.Constraints[0].Value), &members); err != nil {
			http.Error(response, "invalid members", http.StatusBadRequest)
			return true
		}
		for _, member := range members {
			state.members[member] = struct{}{}
		}
	}
	state.resources[identity] = resource.Payload
	counts := f.creates[target]
	counts.Resources++
	f.creates[target] = counts
	_ = json.NewEncoder(response).Encode(map[string]any{"resource": resource})
	return true
}
