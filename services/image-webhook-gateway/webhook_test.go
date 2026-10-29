package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
			"name": "example/app",
			"package_type": "container",
			"owner": {"login": "example"},
			"package_version": {
				"container_metadata": {"tag": {"name": "prod"}}
			}
		}
	}`)

	event, err := parsePackageEvent(body)
	require.NoError(t, err)
	assert.Equal(t, "example/app", event.Package.Name)
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

const trackedImageUpdater = `{
	"spec": {
		"applicationRefs": [
			{
				"namePattern": "tadoku-api",
				"images": [{"alias": "tadoku-api", "imageName": "ghcr.io/tadoku/tadoku/tadoku-api:prod"}]
			},
			{
				"namePattern": "tadoku-frontend-authv2",
				"images": [{"alias": "frontend-authv2", "imageName": "ghcr.io/tadoku/tadoku/frontend-authv2:latest"}]
			}
		]
	}
}`

func TestHandleWebhookQueuesOnlyTrackedImageTags(t *testing.T) {
	tests := []struct {
		name           string
		pkg            string
		packageVersion string
		queued         bool
	}{
		{
			name:           "tracked prod tag",
			pkg:            "tadoku/tadoku-api",
			packageVersion: `{"container_metadata":{"tag":{"name":"prod"}}}`,
			queued:         true,
		},
		{
			name:           "tracked latest tag",
			pkg:            "tadoku/frontend-authv2",
			packageVersion: `{"container_metadata":{"tag":{"name":"latest"}}}`,
			queued:         true,
		},
		{
			name:           "untracked tag of a tracked image",
			pkg:            "tadoku/tadoku-api",
			packageVersion: `{"container_metadata":{"tag":{"name":"latest"}}}`,
		},
		{
			name:           "commit tag",
			pkg:            "tadoku/frontend-authv2",
			packageVersion: `{"container_metadata":{"tag":{"name":"3a7abe859a2ba0b5c0f5c4bb5a5e1d0f1e2a3b4c"}}}`,
		},
		{
			name:           "untagged version",
			pkg:            "tadoku/frontend-authv2",
			packageVersion: `{"name":"sha256:472d63bc","container_metadata":{"tag":{"name":""}}}`,
		},
		{
			name:           "untracked image",
			pkg:            "tadoku/frontend-styleguide",
			packageVersion: `{"container_metadata":{"tag":{"name":"prod"}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := newTestGateway(staticImageUpdater{document: trackedImageUpdater})
			body := publishedPackageEvent(test.pkg, test.packageVersion)

			response := sendWebhook(g, body)

			if response.Code != http.StatusAccepted {
				t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusAccepted, response.Body)
			}
			if !test.queued {
				if len(g.queue) != 0 {
					t.Errorf("queued %d events, want none", len(g.queue))
				}
				return
			}
			if len(g.queue) != 1 {
				t.Fatalf("queued %d events, want 1", len(g.queue))
			}
			if queued := <-g.queue; !bytes.Equal(queued.body, body) {
				t.Errorf("queued body = %s, want the received body", queued.body)
			}
		})
	}
}

func TestHandleWebhookRefusesEventWhenImageUpdaterIsUnreadable(t *testing.T) {
	g := newTestGateway(staticImageUpdater{err: errors.New("Kubernetes API unavailable")})
	body := publishedPackageEvent("tadoku/tadoku-api", `{"container_metadata":{"tag":{"name":"prod"}}}`)

	response := sendWebhook(g, body)

	if response.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if len(g.queue) != 0 {
		t.Errorf("queued %d events, want none", len(g.queue))
	}
}

func newTestGateway(kube imageUpdaterReader) *gateway {
	g := &gateway{
		cfg:   Config{GHCRWebhookSecret: "secret"},
		kube:  kube,
		queue: make(chan webhookEvent, 1),
	}
	g.accepting.Store(true)
	return g
}

func publishedPackageEvent(pkg, packageVersion string) []byte {
	return []byte(`{
		"action": "published",
		"package": {
			"name": "` + pkg + `",
			"package_type": "CONTAINER",
			"owner": {"login": "tadoku"},
			"package_version": ` + packageVersion + `
		}
	}`)
}

func sendWebhook(g *gateway, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/webhook?type=ghcr.io", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "package")
	req.Header.Set("X-Hub-Signature-256", signature([]byte("secret"), body))
	response := httptest.NewRecorder()
	g.handleWebhook(response, req)
	return response
}

type staticImageUpdater struct {
	document string
	err      error
}

func (s staticImageUpdater) get(context.Context) (*imageUpdater, error) {
	if s.err != nil {
		return nil, s.err
	}
	var resource imageUpdater
	if err := json.Unmarshal([]byte(s.document), &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

func TestHandleWebhookRejectsInvalidSignature(t *testing.T) {
	g := &gateway{
		cfg:   Config{GHCRWebhookSecret: "secret"},
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
