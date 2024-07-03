package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/gzip"
	"github.com/maxpain/shortener/internal/handler"
	"github.com/maxpain/shortener/internal/log"
)

type Application struct {
	handler *handler.Handler
	Router  *chi.Mux
}

func NewApplication(
	cfg *config.Configuration,
	logger *log.Logger,
) (*Application, error) {
	handler := handler.NewHandler(cfg, logger)
	router := initRouter(handler, logger)

	return &Application{handler: handler, Router: router}, nil
}

func (a *Application) Close() error {
	return a.handler.Close()
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
