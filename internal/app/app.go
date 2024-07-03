package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/db"
	"github.com/maxpain/shortener/internal/gzip"
	"github.com/maxpain/shortener/internal/handler"
	"github.com/maxpain/shortener/internal/log"
	"github.com/maxpain/shortener/internal/storage"
)

type Application struct {
	Router *chi.Mux
}

func NewApplication(
	cfg *config.Configuration,
	logger *log.Logger,
	DB *db.Database,
) (*Application, error) {
	storage, err := storage.NewStorage(cfg, logger)

	if err != nil {
		return nil, err
	}

	handler := handler.NewHandler(storage, logger, DB)
	router := initRouter(handler, logger)

	return &Application{Router: router}, nil
}

func initRouter(handler *handler.Handler, logger *log.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(logger.Middleware)
	r.Use(gzip.Middleware)

	r.Get("/ping", handler.Ping)
	r.Get("/{hash}", handler.Redirect)
	r.Post("/", handler.Shorten)
	r.Post("/api/shorten", handler.ShortenJSON)
	r.NotFound(handler.NotFound)
	r.MethodNotAllowed(handler.NotFound)

	return r
}
