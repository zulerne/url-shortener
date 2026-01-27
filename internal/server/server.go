package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	HttpServer      *http.Server
	ShutdownTimeout time.Duration
}

func (s *Server) Listen(ctx context.Context) error {
	slog.Info("Starting server", "port", s.HttpServer.Addr)

	errChan := make(chan error, 1)
	go func() {
		err := s.HttpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		slog.Info("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
		defer cancel()
		return s.HttpServer.Shutdown(shutdownCtx)
	}
}
