package config

import (
	"flag"
)

var (
	ServerAddr = flag.String("a", ":8080", "Server address")
	BaseURL    = flag.String("b", "http://localhost:8080", "Base url for generated links")
)

func Init() {
	flag.Parse()
}
