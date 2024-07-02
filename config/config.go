package config

import (
	"flag"
	"os"
)

var (
	ServerAddr      = flag.String("a", ":8080", "Server address")
	BaseURL         = flag.String("b", "http://localhost:8080", "Base url for generated links")
	FileStoragePath = flag.String("f", "/tmp/short-url-db.json", "Path to file storage")
)

func Init() {
	flag.Parse()

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		*ServerAddr = envServerAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		*BaseURL = envBaseURL
	}

	if envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		*FileStoragePath = envFileStoragePath
	}
}
