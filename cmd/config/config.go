package config

import (
	"flag"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func GetConfig() *Config {
	var serverAddress, baseURL string

	flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&baseURL, "b", "http://localhost:8080/", "Base address for shortened URLs")
	flag.Parse()

	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
	}
}
