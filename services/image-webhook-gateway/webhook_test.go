package main

import (
	"bytes"
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
