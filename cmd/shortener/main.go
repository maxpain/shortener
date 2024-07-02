package main

import (
	"net/http"

	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/app"
	"github.com/maxpain/shortener/internal/logger"
)

func main() {
	config.Init()
	logger.Init()

	app, err := app.NewApp(*config.FileStoragePath)

	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(*config.ServerAddr, app.Router)

	if err != nil {
		panic(err)
	}
}
