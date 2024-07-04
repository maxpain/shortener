package main

import (
	"context"
	stdlog "log"
	"net/http"

	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/app"
	"github.com/maxpain/shortener/internal/log"
)

func main() {
	logger, err := log.NewLogger()

	if err != nil {
		stdlog.Fatalf("Error creating logger: %v", err)
	}

	cfg := config.NewConfiguration()
	cfg.ParseFlags()
	cfg.LoadFromEnv()

	app, err := app.NewApplication(context.Background(), cfg, logger)

	if err != nil {
		logger.Sugar().Fatalf("Error creating app: %v", err)
	}

	defer app.Close()

	if err = http.ListenAndServe(cfg.ServerAddr, app.Router); err != nil {
		logger.Sugar().Fatalf("Error starting server: %v", err)
	}
}
