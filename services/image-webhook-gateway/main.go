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

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address                string `required:"true"`
	ImageUpdaterWebhookURL string `envconfig:"image_updater_webhook_url" required:"true"`
	ImageUpdaterNamespace  string `envconfig:"image_updater_namespace" required:"true"`
	ImageUpdaterName       string `envconfig:"image_updater_name" required:"true"`
	QueueSize              int    `envconfig:"queue_size" required:"true"`
	GHCRWebhookSecret      string `envconfig:"ghcr_webhook_secret" required:"true"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	kube, err := newKubeClient(cfg.ImageUpdaterNamespace, cfg.ImageUpdaterName)
	if err != nil {
		slog.Error("configure Kubernetes client", "error", err)
		os.Exit(1)
	}

	g := &gateway{
		cfg:   cfg,
		http:  &http.Client{Timeout: 30 * time.Second},
		kube:  kube,
		queue: make(chan webhookEvent, cfg.QueueSize),
	}
	g.accepting.Store(true)

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", g.handleWebhook)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/readyz", g.handleReady)
	server := &http.Server{
		Addr:              cfg.Address,
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
		slog.Info("image webhook gateway listening", "address", cfg.Address)
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

func loadConfig() (Config, error) {
	cfg := Config{}
	if err := envconfig.Process("GATEWAY", &cfg); err != nil {
		return Config{}, err
	}
	if _, err := url.ParseRequestURI(cfg.ImageUpdaterWebhookURL); err != nil {
		return Config{}, fmt.Errorf("GATEWAY_IMAGE_UPDATER_WEBHOOK_URL: %w", err)
	}
	return cfg, nil
}
