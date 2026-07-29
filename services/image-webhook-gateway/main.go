package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type config struct {
	address      string
	targetURL    string
	namespace    string
	imageUpdater string
	queueSize    int
	secret       string
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_ = server.Shutdown(shutdownCtx)
	cancel()
	close(g.queue)
	workers.Wait()
	slog.Info("queue drained; shutdown complete")
}

func loadConfig() (config, error) {
	cfg := config{
		address:      envOrDefault("ADDRESS", ":8080"),
		targetURL:    envOrDefault("IMAGE_UPDATER_WEBHOOK_URL", "http://argocd-image-updater-webhook-internal:8082/webhook?type=ghcr.io"),
		namespace:    envOrDefault("IMAGE_UPDATER_NAMESPACE", "argocd"),
		imageUpdater: envOrDefault("IMAGE_UPDATER_NAME", "tadoku"),
		queueSize:    256,
		secret:       os.Getenv("GHCR_WEBHOOK_SECRET"),
	}
	if cfg.secret == "" {
		return config{}, errors.New("GHCR_WEBHOOK_SECRET is required")
	}
	if _, err := url.ParseRequestURI(cfg.targetURL); err != nil {
		return config{}, fmt.Errorf("IMAGE_UPDATER_WEBHOOK_URL: %w", err)
	}
	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
