package main

import (
	"net/http"
	//"go.uber.org/zap"
    
	"github.com/ma-shulgin/go-link-shortener/cmd/config"
	"github.com/ma-shulgin/go-link-shortener/internal/app"
  "github.com/ma-shulgin/go-link-shortener/internal/logger"
)

func main() {
	cfg := config.GetConfig()

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		panic(err)
	}

	urlStorage := make(map[string]string)

	logger.Log.Infow("Starting server", "address", cfg.ServerAddress)
	err := http.ListenAndServe(cfg.ServerAddress, app.RootRouter(urlStorage, cfg.BaseURL))
	if err != nil {
		panic(err)
	}
}
