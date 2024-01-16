package main

import (
	"net/http"

	"github.com/ma-shulgin/go-link-shortener/cmd/config"
	"github.com/ma-shulgin/go-link-shortener/internal/app"
	"github.com/ma-shulgin/go-link-shortener/internal/logger"
)

func main() {
	cfg := config.GetConfig()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	logger.Log.Debugln("Parsed config:", cfg)

	urlStorage, err := app.InitURLStore(cfg.FileStoragePath)
	if err != nil {
		logger.Log.Fatal(err)
	}
	defer urlStorage.Close()

	logger.Log.Infow("Starting server", "address", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, app.RootRouter(urlStorage, cfg.BaseURL))
	if err != nil {
		logger.Log.Fatal(err)
	}
}
