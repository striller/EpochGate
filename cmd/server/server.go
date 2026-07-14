package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type server struct {
	addr    string
	handler http.Handler
}

func (s *server) ListenAndServe(ctx context.Context) error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("server started", "addr", s.addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
