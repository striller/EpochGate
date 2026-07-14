package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/EpochGate/internal/config"
	"github.com/EpochGate/internal/proxy"
	"github.com/EpochGate/internal/router"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	proxyHandler, err := proxy.New(cfg)
	if err != nil {
		slog.Error("failed to create proxy", "error", err)
		os.Exit(1)
	}

	r := router.New(proxyHandler)

	srv := &server{
		addr:    cfg.ListenPort,
		handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.ListenAndServe(ctx); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
