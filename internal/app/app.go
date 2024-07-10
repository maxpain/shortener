package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/handler"
	memoryRepository "github.com/maxpain/shortener/internal/repository/memory"
	postgresRepository "github.com/maxpain/shortener/internal/repository/postgres"
	"github.com/maxpain/shortener/internal/usecase"
)

type App struct {
	*fiber.App
	logger     *slog.Logger
	repository usecase.Repository
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
	repo, err := getRepository(ctx, cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	err = repo.Init(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	useCase := usecase.New(repo, logger)
	handler := handler.New(useCase, logger, cfg.BaseURL)
	app := fiber.New()
	setupRoutes(app, cfg, logger, handler)

	return &App{
		App:        app,
		logger:     logger,
		repository: repo,
	}, nil
}

func getRepository(ctx context.Context, cfg *config.Config, logger *slog.Logger) (usecase.Repository, error) {
	if cfg.DatabaseDSN != "" {
		db, err := pgxpool.New(ctx, cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		logger.Info("Initialized postgres repository")

		return postgresRepository.New(db, logger), nil
	}

	var file *os.File

	if cfg.FileStoragePath != "" {
		var err error

		file, err = os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s : %w", cfg.FileStoragePath, err)
		}

		logger.Info("Initialized memory repository with persistence")
	} else {
		logger.Info("Initialized memory repository without persistence")
	}

	return memoryRepository.New(file, logger), nil
}

func (a *App) Close() {
	a.repository.Close()
}
