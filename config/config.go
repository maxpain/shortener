package config

import (
	"flag"
	"os"
)

type Configuration struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func NewConfiguration() *Configuration {
	return &Configuration{
		ServerAddr:      ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/tmp/short-url-db.json",
		DatabaseDSN:     "",
	}
}

func (c *Configuration) ParseFlags() {
	flag.StringVar(&c.ServerAddr, "a", c.ServerAddr, "Server address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base url for generated links")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Path to file storage")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Database DSN (optional)")

	flag.Parse()

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		c.ServerAddr = envServerAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		c.BaseURL = envBaseURL
	}

	if envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		c.FileStoragePath = envFileStoragePath
	}

	if envDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		c.DatabaseDSN = envDatabaseDSN
	}
}
