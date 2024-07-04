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

type Option func(*Configuration)

func NewConfiguration(opts ...Option) *Configuration {
	cfg := &Configuration{
		ServerAddr:      ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/tmp/short-url-db.json",
		DatabaseDSN:     "",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func WithServerAddr(addr string) Option {
	return func(c *Configuration) {
		c.ServerAddr = addr
	}
}

func WithBaseURL(url string) Option {
	return func(c *Configuration) {
		c.BaseURL = url
	}
}

func WithFileStoragePath(path string) Option {
	return func(c *Configuration) {
		c.FileStoragePath = path
	}
}

func WithDatabaseDSN(dsn string) Option {
	return func(c *Configuration) {
		c.DatabaseDSN = dsn
	}
}

func (c *Configuration) ParseFlags() {
	flag.StringVar(&c.ServerAddr, "a", c.ServerAddr, "Server address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base url for generated links")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Path to file storage")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Database DSN (optional)")

	flag.Parse()
}

func (c *Configuration) LoadFromEnv() {
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		c.ServerAddr = addr
	}

	if url := os.Getenv("BASE_URL"); url != "" {
		c.BaseURL = url
	}

	if path, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		c.FileStoragePath = path
	}

	if dsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		c.DatabaseDSN = dsn
	}
}
