package main

import (
	stdlog "log"
	"net/http"

	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/app"
	"github.com/maxpain/shortener/internal/db"
	"github.com/maxpain/shortener/internal/log"
)

func main() {
	logger, err := log.NewLogger()

	if err != nil {
		stdlog.Fatalf("Error creating logger: %v", err)
	}

	cfg := config.NewConfiguration()
	cfg.ParseFlags()

	DB, err := db.NewDatabase(cfg)

	if err != nil {
		logger.Sugar().Fatalf("Error creating database: %v", err)
	}

	defer DB.Close()

	app, err := app.NewApplication(cfg, logger, DB)

	if err != nil {
		logger.Sugar().Fatalf("Error creating app: %v", err)
	}

	if err = http.ListenAndServe(cfg.ServerAddr, app.Router); err != nil {
		logger.Sugar().Fatalf("Error starting server: %v", err)
	}
}
