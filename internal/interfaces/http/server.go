package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Server оборачивает http.Server для корректного shutdown.
type Server struct {
	http *http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		http: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) ListenAndServe() error {
	err := s.http.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
