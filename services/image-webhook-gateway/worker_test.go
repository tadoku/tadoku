package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
