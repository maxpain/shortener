package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/maxpain/shortener/internal/gzip"
	"github.com/maxpain/shortener/internal/handler"
	"github.com/maxpain/shortener/internal/logger"
	"github.com/maxpain/shortener/internal/storage"
)

type App struct {
	Router *chi.Mux
}

func NewApp() *App {
	storage := storage.NewStorage()
	handler := handler.NewHandler(storage)

	r := chi.NewRouter()

	r.Use(logger.Middleware)
	r.Use(gzip.Middleware)

	r.Get("/{hash}", handler.Redirect)
	r.Post("/", handler.Shorten)
	r.Post("/api/shorten", handler.ShortenJSON)
	r.NotFound(handler.NotFound)
	r.MethodNotAllowed(handler.NotFound)

	return &App{Router: r}
}
