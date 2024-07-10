package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/app"
)

func main() {
	cfg := config.New()
	cfg.ParseFlags()
	cfg.LoadFromEnv()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	shortenerApp, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("failed to create app", slog.Any("error", err))
		os.Exit(1)
	}

	defer shortenerApp.Close()

	err = shortenerApp.Listen(cfg.ServerAddr)
	if err != nil {
		logger.Error("failed to start app", slog.Any("error", err))
	}
}
