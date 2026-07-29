package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignatureValidation(t *testing.T) {
	body := []byte(`{"action":"published"}`)
	secret := []byte("secret")

	assert.True(t, validSignature(secret, body, signature(secret, body)))
	assert.False(t, validSignature(secret, body, "sha256=wrong"))
	assert.False(t, validSignature(secret, []byte("changed"), signature(secret, body)))
}

func TestParsePackageEvent(t *testing.T) {
	body := []byte(`{
		"action": "published",
		"package": {
			"name": "tadoku/content-api",
			"package_type": "container",
			"owner": {"login": "tadoku"},
			"package_version": {
				"container_metadata": {"tag": {"name": "prod"}}
			}
		}
	}`)

	event, err := parsePackageEvent(body)
	require.NoError(t, err)
	assert.Equal(t, "tadoku/content-api", event.Package.Name)
}

func TestParsePackageEventRejectsUnusableEvents(t *testing.T) {
	tests := []string{
		`{`,
		`{"action":"deleted"}`,
		`{"action":"published","package":{"package_type":"npm"}}`,
		`{"action":"published","package":{"package_type":"container","name":"app","owner":{"login":"org"}}}`,
	}
	for _, body := range tests {
		_, err := parsePackageEvent([]byte(body))
		assert.Error(t, err)
	}
}

func TestHandleWebhookQueuesAuthenticatedEvent(t *testing.T) {
	g := &gateway{
		cfg:   config{secret: "secret"},
		queue: make(chan webhookEvent, 1),
	}
	g.accepting.Store(true)
	body := []byte(`{
		"action":"published",
		"package":{
			"name":"tadoku/content-api",
			"package_type":"container",
			"owner":{"login":"tadoku"},
			"package_version":{"name":"prod"}
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook?type=ghcr.io", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "package")
	req.Header.Set("X-Hub-Signature-256", signature([]byte("secret"), body))
	response := httptest.NewRecorder()

	g.handleWebhook(response, req)

	assert.Equal(t, http.StatusAccepted, response.Code)
	queued := <-g.queue
	assert.Equal(t, body, queued.body)
}

func TestHandleWebhookIgnoresAuthenticatedNonProdEvent(t *testing.T) {
	g := &gateway{
		cfg:   config{secret: "secret"},
		queue: make(chan webhookEvent, 1),
	}
	g.accepting.Store(true)
	body := []byte(`{
		"action":"published",
		"package":{
			"name":"tadoku/content-api",
			"package_type":"container",
			"owner":{"login":"tadoku"},
			"package_version":{"name":"latest"}
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "package")
	req.Header.Set("X-Hub-Signature-256", signature([]byte("secret"), body))
	response := httptest.NewRecorder()

	g.handleWebhook(response, req)

	assert.Equal(t, http.StatusAccepted, response.Code)
	assert.Empty(t, g.queue)
}

func TestHandleWebhookRejectsInvalidSignature(t *testing.T) {
	g := &gateway{
		cfg:   config{secret: "secret"},
		queue: make(chan webhookEvent, 1),
	}
	g.accepting.Store(true)
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-GitHub-Event", "package")
	req.Header.Set("X-Hub-Signature-256", "sha256=wrong")
	response := httptest.NewRecorder()

	g.handleWebhook(response, req)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Empty(t, g.queue)
}

func TestWorkerSerializesUntilReconciliationCompletes(t *testing.T) {
	kube := &fakeImageUpdaterReader{lastCheckedAt: "initial"}
	var mu sync.Mutex
	active := 0
	maxActive := 0
	completed := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		active++
		if active > maxActive {
			maxActive = active
		}
		mu.Unlock()
		kube.setReconciling(true)
		go func() {
			time.Sleep(20 * time.Millisecond)
			mu.Lock()
			active--
			completed++
			value := completed
			mu.Unlock()
			kube.complete(value)
		}()
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	g := &gateway{
		cfg:       config{targetURL: upstream.URL, secret: "secret"},
		http:      upstream.Client(),
		kube:      kube,
		queue:     make(chan webhookEvent, 2),
		pollEvery: time.Millisecond,
	}
	g.queue <- webhookEvent{body: []byte(`{}`)}
	g.queue <- webhookEvent{body: []byte(`{}`)}
	close(g.queue)

	g.runWorker()

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 2, completed)
	assert.Equal(t, 1, maxActive)
}

type fakeImageUpdaterReader struct {
	mu            sync.Mutex
	lastCheckedAt string
	reconciling   bool
}

func (f *fakeImageUpdaterReader) get(context.Context) (*imageUpdater, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	resource := &imageUpdater{}
	resource.Status.LastCheckedAt = f.lastCheckedAt
	status := "False"
	if f.reconciling {
		status = "True"
	}
	resource.Status.Conditions = append(resource.Status.Conditions, struct {
		Type    string `json:"type"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
	}{Type: "Reconciling", Status: status})
	return resource, nil
}

func (f *fakeImageUpdaterReader) setReconciling(value bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reconciling = value
}

func (f *fakeImageUpdaterReader) complete(sequence int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastCheckedAt = time.Unix(int64(sequence), 0).UTC().Format(time.RFC3339Nano)
	f.reconciling = false
}
