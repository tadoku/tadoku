package flipt

import (
	"context"
	"fmt"
	"strings"
	"sync"

	featureflags "github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	fliptsdk "go.flipt.io/flipt-client"
)

type fakeSDKClient struct {
	response *fliptsdk.BooleanEvaluationResponse
	err      error
	stateErr error
	request  *fliptsdk.EvaluationRequest
	closed   bool
}

func (c *fakeSDKClient) EvaluateBoolean(_ context.Context, request *fliptsdk.EvaluationRequest) (*fliptsdk.BooleanEvaluationResponse, error) {
	c.request = request
	return c.response, c.err
}

func (c *fakeSDKClient) Err() error { return c.stateErr }

func (c *fakeSDKClient) Close(_ context.Context) error {
	c.closed = true
	return nil
}

type recordingObserver struct {
	mu        sync.Mutex
	statuses  []featureflags.InitializationStatus
	refreshes int
	errors    []string
}

func (o *recordingObserver) ObserveInitialization(status featureflags.InitializationStatus) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.statuses = append(o.statuses, status)
}

func (o *recordingObserver) ObserveConfigRefresh() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.refreshes++
}
func (o *recordingObserver) ObserveProviderError(kind string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.errors = append(o.errors, kind)
}

func (o *recordingObserver) snapshot() ([]featureflags.InitializationStatus, int, []string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]featureflags.InitializationStatus(nil), o.statuses...), o.refreshes, append([]string(nil), o.errors...)
}

func booleanSnapshot(rollouts []string) string {
	if rollouts == nil {
		rollouts = []string{}
	}
	return fmt.Sprintf(`{
        "namespace": {"key": "default"},
        "flags": [{
			"key": "fixture-flag",
			"name": "",
			"description": "",
			"enabled": false,
			"type": "BOOLEAN_FLAG_TYPE",
			"createdAt": "2026-08-24T06:01:42Z",
			"updatedAt": "2026-08-24T06:01:42Z",
			"rules": [],
			"rollouts": [%s]
		}],
		"digest": "f86fe0149bc542c8b350e994b869ee5370411c12"
    }`, strings.Join(rollouts, ","))
}
