package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr      string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	JwtSecret       string
}

type Option func(*Config)

func New(opts ...Option) *Config {
	cfg := &Config{
		ServerAddr:      ":8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/tmp/short-url-db.json",
		DatabaseDSN:     "",
		JwtSecret:       "secret",
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

func WithServerAddr(addr string) Option {
	return func(c *Config) {
		c.ServerAddr = addr
	}
}

func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

func WithFileStoragePath(path string) Option {
	return func(c *Config) {
		c.FileStoragePath = path
	}
}

func WithDatabaseDSN(dsn string) Option {
	return func(c *Config) {
		c.DatabaseDSN = dsn
	}
}

func WithJwtSecret(secret string) Option {
	return func(c *Config) {
		c.JwtSecret = secret
	}
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.ServerAddr, "a", c.ServerAddr, "Server address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base url for generated links")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Path to file storage")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Database DSN (optional)")
	flag.StringVar(&c.JwtSecret, "j", c.JwtSecret, "JWT secret")

	flag.Parse()
}

func (c *Config) LoadFromEnv() {
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

	if secret, ok := os.LookupEnv("JWT_SECRET"); ok {
		c.JwtSecret = secret
	}
}
