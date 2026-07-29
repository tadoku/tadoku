package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

const maxWebhookBody = 1 << 20

type webhookEvent struct {
	body     []byte
	delivery string
}

type packageEvent struct {
	Action  string `json:"action"`
	Package struct {
		Name        string `json:"name"`
		PackageType string `json:"package_type"`
		Owner       struct {
			Login string `json:"login"`
		} `json:"owner"`
		PackageVersion struct {
			Version           string `json:"version,omitempty"`
			Name              string `json:"name,omitempty"`
			ContainerMetadata struct {
				Tag struct {
					Name string `json:"name,omitempty"`
				} `json:"tag,omitempty"`
			} `json:"container_metadata,omitempty"`
		} `json:"package_version"`
	} `json:"package"`
}

func (g *gateway) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if !g.accepting.Load() {
		http.Error(w, "shutting down", http.StatusServiceUnavailable)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method must be POST", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("X-GitHub-Event") != "package" {
		http.Error(w, "unsupported GitHub event", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBody))
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if !validSignature([]byte(g.cfg.GHCRWebhookSecret), body, r.Header.Get("X-Hub-Signature-256")) {
		http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
		return
	}
	eventPayload, err := parsePackageEvent(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if eventPayload.tag() != "prod" {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("ignored non-prod package event"))
		return
	}

	event := webhookEvent{
		body:     body,
		delivery: r.Header.Get("X-GitHub-Delivery"),
	}
	select {
	case g.queue <- event:
		slog.Info("webhook event queued", "delivery", event.delivery, "queued", len(g.queue))
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("queued"))
	default:
		http.Error(w, "event queue is full", http.StatusServiceUnavailable)
	}
}

func (g *gateway) handleReady(w http.ResponseWriter, _ *http.Request) {
	if !g.accepting.Load() {
		http.Error(w, "not accepting events", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func parsePackageEvent(body []byte) (packageEvent, error) {
	var event packageEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return event, errors.New("invalid JSON payload")
	}
	if event.Action != "published" {
		return event, fmt.Errorf("unsupported package action %q", event.Action)
	}
	if !strings.EqualFold(event.Package.PackageType, "container") {
		return event, fmt.Errorf("unsupported package type %q", event.Package.PackageType)
	}
	if event.Package.Name == "" || event.Package.Owner.Login == "" {
		return event, errors.New("package owner and name are required")
	}
	if event.tag() == "" {
		return event, errors.New("package tag is required")
	}
	return event, nil
}

func (e packageEvent) tag() string {
	if tag := e.Package.PackageVersion.ContainerMetadata.Tag.Name; tag != "" {
		return tag
	}
	if tag := e.Package.PackageVersion.Name; tag != "" {
		return tag
	}
	return e.Package.PackageVersion.Version
}

func signature(secret, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func validSignature(secret, body []byte, supplied string) bool {
	expected := signature(secret, body)
	return hmac.Equal([]byte(expected), []byte(supplied))
}
