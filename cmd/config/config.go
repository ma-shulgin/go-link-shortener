package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func GetConfig() *Config {
	var serverAddress, baseURL string

	flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&baseURL, "b", "http://localhost:8080", "Base address for shortened URLs")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		serverAddress = envServerAddress
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}

	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
	}
}
