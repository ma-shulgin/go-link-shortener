package main

import (
	"net/http"

	"github.com/ma-shulgin/go-link-shortener/cmd/config"
	"github.com/ma-shulgin/go-link-shortener/internal/app"
	"github.com/ma-shulgin/go-link-shortener/internal/logger"
	"github.com/ma-shulgin/go-link-shortener/internal/storage"
)

func main() {
	cfg := config.GetConfig()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	logger.Log.Debugln("Parsed config:", cfg)

	var urlStore storage.URLStore
	var err error
	if cfg.DatabaseDSN != "" {
		urlStore, err = storage.InitPostgresStore(cfg.DatabaseDSN)
	} else if cfg.FileStoragePath != "" {
		urlStore, err = storage.InitFileStore(cfg.FileStoragePath)
	} else {
		urlStore = storage.InitMemoryStore()
	}

	if err != nil {
		logger.Log.Fatal(err)
	}
	defer urlStore.Close()

	logger.Log.Infow("Starting server", "address", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, app.RootRouter(urlStore, cfg.BaseURL))
	if err != nil {
		logger.Log.Fatal(err)
	}
}
