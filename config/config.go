package config

import (
	"flag"
	"os"
)

var (
	ServerAddr = flag.String("a", ":8080", "Server address")
	BaseURL    = flag.String("b", "http://localhost:8080", "Base url for generated links")
)

func Init() {
	flag.Parse()

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		*ServerAddr = envServerAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		*BaseURL = envBaseURL
	}
}
