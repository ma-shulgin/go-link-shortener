package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	LogLevel        string
	FileStoragePath string
}

func GetConfig() *Config {
	var serverAddress, baseURL, logLevel, fileStoragePath string

	flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&baseURL, "b", "http://localhost:8080", "Base address for shortened URLs")
	flag.StringVar(&logLevel, "l", "info", "log level")
	flag.StringVar(&fileStoragePath, "f", "/tmp/short-url-db.json", "File storage path")
	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		serverAddress = envServerAddress
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		baseURL = envBaseURL
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		logLevel = envLogLevel
	}
	if envfileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envfileStoragePath != "" {
		fileStoragePath = envfileStoragePath
	}

	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
		LogLevel:      logLevel,
		FileStoragePath: fileStoragePath,
	}
}
