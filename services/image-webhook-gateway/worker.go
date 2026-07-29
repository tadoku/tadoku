package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

const (
	retryDelay = 5 * time.Second
	statusPoll = time.Second
)

type gateway struct {
	cfg       Config
	http      *http.Client
	kube      imageUpdaterReader
	queue     chan webhookEvent
	accepting atomic.Bool
	pollEvery time.Duration
}

func (g *gateway) runWorker() {
	for event := range g.queue {
		g.process(event)
	}
}

func (g *gateway) process(event webhookEvent) {
	for {
		baseline, err := g.waitUntilIdle(context.Background())
		if err != nil {
			slog.Error("wait for ImageUpdater idle state", "error", err)
			time.Sleep(retryDelay)
			continue
		}
		if err := g.forward(context.Background(), event); err != nil {
			slog.Error("forward webhook", "delivery", event.delivery, "error", err)
			time.Sleep(retryDelay)
			continue
		}
		status, err := g.waitForCompletion(context.Background(), baseline)
		if err != nil {
			slog.Error("wait for webhook reconciliation", "delivery", event.delivery, "error", err)
			time.Sleep(retryDelay)
			continue
		}
		slog.Info("serialized webhook reconciliation completed",
			"delivery", event.delivery,
			"lastCheckedAt", status.Status.LastCheckedAt,
			"readyReason", conditionReason(status, "Ready"),
		)
		return
	}
}

func (g *gateway) forward(ctx context.Context, event webhookEvent) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.cfg.ImageUpdaterWebhookURL, bytes.NewReader(event.body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "package")
	req.Header.Set("X-Hub-Signature-256", signature([]byte(g.cfg.GHCRWebhookSecret), event.body))
	if event.delivery != "" {
		req.Header.Set("X-GitHub-Delivery", event.delivery)
	}
	resp, err := g.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("Image Updater returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func (g *gateway) waitUntilIdle(ctx context.Context) (string, error) {
	for {
		resource, err := g.kube.get(ctx)
		if err != nil {
			slog.Error("read ImageUpdater status while waiting for idle", "error", err)
			if err := wait(ctx, retryDelay); err != nil {
				return "", err
			}
			continue
		}
		if !conditionTrue(resource, "Reconciling") {
			return resource.Status.LastCheckedAt, nil
		}
		if err := wait(ctx, g.statusPollInterval()); err != nil {
			return "", err
		}
	}
}

func (g *gateway) waitForCompletion(ctx context.Context, baseline string) (*imageUpdater, error) {
	for {
		resource, err := g.kube.get(ctx)
		if err != nil {
			slog.Error("read ImageUpdater status while waiting for completion", "error", err)
			if err := wait(ctx, retryDelay); err != nil {
				return nil, err
			}
			continue
		}
		if resource.Status.LastCheckedAt != "" &&
			resource.Status.LastCheckedAt != baseline &&
			!conditionTrue(resource, "Reconciling") {
			return resource, nil
		}
		if err := wait(ctx, g.statusPollInterval()); err != nil {
			return nil, err
		}
	}
}

func (g *gateway) statusPollInterval() time.Duration {
	if g.pollEvery > 0 {
		return g.pollEvery
	}
	return statusPoll
}

func conditionTrue(resource *imageUpdater, conditionType string) bool {
	for _, condition := range resource.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Status == "True"
		}
	}
	return false
}

func conditionReason(resource *imageUpdater, conditionType string) string {
	for _, condition := range resource.Status.Conditions {
		if condition.Type == conditionType {
			return condition.Reason
		}
	}
	return ""
}

func wait(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
