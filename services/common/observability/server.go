package observability

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	server   *http.Server
	mu       sync.RWMutex
	listener net.Listener
	errors   chan error
}

func NewServer(address string, handler http.Handler) *Server {
	return &Server{
		server: &http.Server{
			Addr:              address,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
		errors: make(chan error, 1),
	}
}

func (s *Server) Start() error {
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

func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *Server) Errors() <-chan error {
	return s.errors
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
