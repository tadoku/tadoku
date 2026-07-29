package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	maxWebhookBody = 1 << 20
	retryDelay     = 5 * time.Second
	statusPoll     = time.Second
)

type config struct {
	address         string
	targetURL       string
	namespace       string
	imageUpdater    string
	refreshInterval time.Duration
	queueSize       int
	secret          string
}

type webhookEvent struct {
	body     []byte
	delivery string
	source   string
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

type imageUpdater struct {
	Spec struct {
		ApplicationRefs []struct {
			Images []struct {
				ImageName string `json:"imageName"`
			} `json:"images"`
		} `json:"applicationRefs"`
	} `json:"spec"`
	Status struct {
		LastCheckedAt string `json:"lastCheckedAt"`
		Conditions    []struct {
			Type    string `json:"type"`
			Status  string `json:"status"`
			Reason  string `json:"reason"`
			Message string `json:"message"`
		} `json:"conditions"`
	} `json:"status"`
}

type kubeClient struct {
	client *http.Client
	url    string
	token  string
}

type imageUpdaterReader interface {
	get(context.Context) (*imageUpdater, error)
}

type gateway struct {
	cfg       config
	http      *http.Client
	kube      imageUpdaterReader
	queue     chan webhookEvent
	accepting atomic.Bool
	pending   atomic.Int64
	pollEvery time.Duration
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	kube, err := newKubeClient(cfg.namespace, cfg.imageUpdater)
	if err != nil {
		slog.Error("configure Kubernetes client", "error", err)
		os.Exit(1)
	}

	g := &gateway{
		cfg:   cfg,
		http:  &http.Client{Timeout: 30 * time.Second},
		kube:  kube,
		queue: make(chan webhookEvent, cfg.queueSize),
	}
	g.accepting.Store(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", g.handleWebhook)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", g.handleReady)
	server := &http.Server{
		Addr:              cfg.address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		g.runWorker()
	}()

	refreshCtx, stopRefresh := context.WithCancel(context.Background())
	var scheduler sync.WaitGroup
	scheduler.Add(1)
	go func() {
		defer scheduler.Done()
		g.runRefreshScheduler(refreshCtx)
	}()

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("image webhook gateway listening", "address", cfg.address)
		serverErrors <- server.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-signals:
		slog.Info("shutdown requested; draining queued events", "signal", sig)
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("webhook server stopped", "error", err)
		}
	}

	g.accepting.Store(false)
	stopRefresh()
	scheduler.Wait()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = server.Shutdown(shutdownCtx)
	cancel()
	close(g.queue)
	workers.Wait()
	slog.Info("queue drained; shutdown complete")
}

func loadConfig() (config, error) {
	cfg := config{
		address:         envOrDefault("ADDRESS", ":8080"),
		targetURL:       envOrDefault("IMAGE_UPDATER_WEBHOOK_URL", "http://argocd-image-updater-webhook-internal:8082/webhook?type=ghcr.io"),
		namespace:       envOrDefault("IMAGE_UPDATER_NAMESPACE", "argocd"),
		imageUpdater:    envOrDefault("IMAGE_UPDATER_NAME", "tadoku"),
		refreshInterval: time.Hour,
		queueSize:       256,
		secret:          os.Getenv("GHCR_WEBHOOK_SECRET"),
	}
	var err error
	if value := os.Getenv("REFRESH_INTERVAL"); value != "" {
		cfg.refreshInterval, err = time.ParseDuration(value)
		if err != nil {
			return config{}, fmt.Errorf("REFRESH_INTERVAL: %w", err)
		}
	}
	if cfg.secret == "" {
		return config{}, errors.New("GHCR_WEBHOOK_SECRET is required")
	}
	if _, err := url.ParseRequestURI(cfg.targetURL); err != nil {
		return config{}, fmt.Errorf("IMAGE_UPDATER_WEBHOOK_URL: %w", err)
	}
	if cfg.refreshInterval <= 0 {
		return config{}, errors.New("REFRESH_INTERVAL must be positive")
	}
	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
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
	if !validSignature([]byte(g.cfg.secret), body, r.Header.Get("X-Hub-Signature-256")) {
		http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
		return
	}
	if _, err := parsePackageEvent(body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event := webhookEvent{
		body:     body,
		delivery: r.Header.Get("X-GitHub-Delivery"),
		source:   "github",
	}
	g.pending.Add(1)
	select {
	case g.queue <- event:
		slog.Info("webhook event queued", "delivery", event.delivery, "pending", g.pending.Load())
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("queued"))
	default:
		g.pending.Add(-1)
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

func (g *gateway) runWorker() {
	for event := range g.queue {
		g.process(event)
		g.pending.Add(-1)
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
			slog.Error("forward webhook", "source", event.source, "delivery", event.delivery, "error", err)
			time.Sleep(retryDelay)
			continue
		}
		status, err := g.waitForCompletion(context.Background(), baseline)
		if err != nil {
			slog.Error("wait for webhook reconciliation", "source", event.source, "delivery", event.delivery, "error", err)
			time.Sleep(retryDelay)
			continue
		}
		slog.Info("serialized webhook reconciliation completed",
			"source", event.source,
			"delivery", event.delivery,
			"lastCheckedAt", status.Status.LastCheckedAt,
			"readyReason", conditionReason(status, "Ready"),
		)
		return
	}
}

func (g *gateway) forward(ctx context.Context, event webhookEvent) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.cfg.targetURL, bytes.NewReader(event.body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "package")
	req.Header.Set("X-Hub-Signature-256", signature([]byte(g.cfg.secret), event.body))
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

func (g *gateway) runRefreshScheduler(ctx context.Context) {
	timer := time.NewTimer(g.cfg.refreshInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			g.enqueueRefresh(ctx)
			timer.Reset(g.cfg.refreshInterval)
		}
	}
}

func (g *gateway) enqueueRefresh(ctx context.Context) {
	resource, err := g.kube.get(ctx)
	if err != nil {
		slog.Error("load images for hourly refresh", "error", err)
		return
	}
	seen := make(map[string]struct{})
	for _, app := range resource.Spec.ApplicationRefs {
		for _, image := range app.Images {
			owner, name, tag, ok := splitGHCRImage(image.ImageName)
			if !ok {
				continue
			}
			key := owner + "/" + name + ":" + tag
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			body, err := syntheticPackageEvent(owner, name, tag)
			if err != nil {
				slog.Error("create hourly refresh event", "image", image.ImageName, "error", err)
				continue
			}
			g.pending.Add(1)
			select {
			case g.queue <- webhookEvent{body: body, source: "hourly-refresh"}:
			default:
				g.pending.Add(-1)
				slog.Error("hourly refresh event dropped because queue is full", "image", image.ImageName)
			}
		}
	}
	slog.Info("hourly image refresh queued", "images", len(seen), "pending", g.pending.Load())
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
	tag := event.Package.PackageVersion.ContainerMetadata.Tag.Name
	if tag == "" {
		tag = event.Package.PackageVersion.Name
	}
	if tag == "" {
		tag = event.Package.PackageVersion.Version
	}
	if tag == "" {
		return event, errors.New("package tag is required")
	}
	return event, nil
}

func syntheticPackageEvent(owner, name, tag string) ([]byte, error) {
	var event packageEvent
	event.Action = "published"
	event.Package.Name = name
	event.Package.PackageType = "container"
	event.Package.Owner.Login = owner
	event.Package.PackageVersion.Name = tag
	event.Package.PackageVersion.ContainerMetadata.Tag.Name = tag
	return json.Marshal(event)
}

func splitGHCRImage(image string) (owner, name, tag string, ok bool) {
	const prefix = "ghcr.io/"
	if !strings.HasPrefix(image, prefix) {
		return "", "", "", false
	}
	reference := strings.TrimPrefix(image, prefix)
	colon := strings.LastIndex(reference, ":")
	slash := strings.Index(reference, "/")
	if slash <= 0 || colon <= slash+1 || colon == len(reference)-1 {
		return "", "", "", false
	}
	return reference[:slash], reference[slash+1 : colon], reference[colon+1:], true
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

func newKubeClient(namespace, name string) (*kubeClient, error) {
	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, fmt.Errorf("read service account token: %w", err)
	}
	caBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err != nil {
		return nil, fmt.Errorf("read Kubernetes CA: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) {
		return nil, errors.New("parse Kubernetes CA")
	}
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")
	if host == "" || port == "" {
		return nil, errors.New("Kubernetes service environment is missing")
	}
	apiURL := fmt.Sprintf(
		"https://%s:%s/apis/argocd-image-updater.argoproj.io/v1alpha1/namespaces/%s/imageupdaters/%s",
		host, port, url.PathEscape(namespace), url.PathEscape(name),
	)
	return &kubeClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					RootCAs:    pool,
				},
			},
		},
		url:   apiURL,
		token: strings.TrimSpace(string(tokenBytes)),
	}, nil
}

func (k *kubeClient) get(ctx context.Context) (*imageUpdater, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+k.token)
	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("Kubernetes API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var resource imageUpdater
	if err := json.NewDecoder(resp.Body).Decode(&resource); err != nil {
		return nil, err
	}
	return &resource, nil
}
