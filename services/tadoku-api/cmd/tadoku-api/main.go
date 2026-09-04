package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"
	commonobservability "github.com/tadoku/tadoku/services/common/observability"
	transporthttp "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

type config struct {
	Port                  int           `validate:"gt=0,lte=65535" default:"8000"`
	MetricsPort           int           `validate:"gt=0,lte=65535" envconfig:"metrics_port" default:"9090"`
	ServiceName           string        `validate:"required" envconfig:"service_name" default:"tadoku-api"`
	AuthzURL              string        `validate:"required" envconfig:"authz_url"`
	ContentURL            string        `validate:"required" envconfig:"content_url"`
	ImmersionURL          string        `validate:"required" envconfig:"immersion_url"`
	ProfileURL            string        `validate:"required" envconfig:"profile_url"`
	DialTimeout           time.Duration `validate:"gt=0" envconfig:"dial_timeout" default:"3s"`
	ResponseHeaderTimeout time.Duration `validate:"gt=0" envconfig:"response_header_timeout" default:"10s"`
	RequestTimeout        time.Duration `validate:"gt=0" envconfig:"request_timeout" default:"30s"`
	IdleTimeout           time.Duration `validate:"gt=0" envconfig:"idle_timeout" default:"30s"`
	ShutdownTimeout       time.Duration `validate:"gt=0" envconfig:"shutdown_timeout" default:"10s"`
}

func loadConfig() (config, error) {
	cfg := config{}
	if err := envconfig.Process("API", &cfg); err != nil {
		return config{}, fmt.Errorf("load config: %w", err)
	}
	if err := validator.New().Struct(cfg); err != nil {
		return config{}, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

type application struct {
	server          *http.Server
	listener        net.Listener
	serverErrors    chan error
	metricsServer   *commonobservability.Server
	shutdownTimeout time.Duration
	transport       *http.Transport
}

func start(cfg config, logger *slog.Logger) (*application, error) {
	logger = logger.With("service", cfg.ServiceName)
	metrics := commonobservability.NewMetrics(nil, cfg.ServiceName)
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: cfg.DialTimeout, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   25,
		IdleConnTimeout:       cfg.IdleTimeout,
		TLSHandshakeTimeout:   cfg.DialTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		ExpectContinueTimeout: time.Second,
	}
	handler, err := transporthttp.NewHandler(transporthttp.Upstreams{
		Authz: cfg.AuthzURL, Content: cfg.ContentURL,
		Immersion: cfg.ImmersionURL, Profile: cfg.ProfileURL,
	}, transport, cfg.RequestTimeout, metrics.Registerer(), logger)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("listen for API requests: %w", err)
	}
	metricsServer := commonobservability.NewServer(fmt.Sprintf("0.0.0.0:%d", cfg.MetricsPort), metrics.Handler())
	if err := metricsServer.Start(); err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("start metrics server: %w", err)
	}

	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.RequestTimeout,
		WriteTimeout:      cfg.RequestTimeout + time.Second,
		IdleTimeout:       cfg.IdleTimeout,
	}
	app := &application{
		server: server, listener: listener, serverErrors: make(chan error, 1),
		metricsServer: metricsServer, shutdownTimeout: cfg.ShutdownTimeout, transport: transport,
	}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.serverErrors <- err
		}
	}()
	logger.Info("tadoku-api started", "address", listener.Addr(), "mode", "proxy")
	return app, nil
}

func (app *application) wait(ctx context.Context) error {
	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-app.serverErrors:
		runErr = fmt.Errorf("API server stopped: %w", runErr)
	case runErr = <-app.metricsServer.Errors():
		runErr = fmt.Errorf("metrics server stopped: %w", runErr)
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), app.shutdownTimeout)
	defer cancel()
	app.transport.CloseIdleConnections()
	return errors.Join(runErr, app.server.Shutdown(shutdownContext), app.metricsServer.Shutdown(shutdownContext))
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}
	app, err := start(cfg, logger)
	if err != nil {
		panic(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := app.wait(ctx); err != nil {
		panic(err)
	}
}
