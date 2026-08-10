package observability

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

type MetricsServer struct {
	server   *http.Server
	mu       sync.RWMutex
	listener net.Listener
	errors   chan error
}

func NewMetricsServer(address string, handler http.Handler) *MetricsServer {
	return &MetricsServer{
		server: &http.Server{
			Addr:              address,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
		errors: make(chan error, 1),
	}
}

func (s *MetricsServer) Start() error {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()
	go func() {
		if serveErr := s.server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			s.errors <- serveErr
		}
	}()
	return nil
}

func (s *MetricsServer) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *MetricsServer) Errors() <-chan error {
	return s.errors
}

func (s *MetricsServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
