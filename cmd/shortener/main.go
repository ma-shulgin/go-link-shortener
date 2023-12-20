package main

import (
	"log"
	"net/http"

	"github.com/ma-shulgin/go-link-shortener/cmd/config"
	"github.com/ma-shulgin/go-link-shortener/internal/app"
)

func main() {
	cfg := config.GetConfig()

	urlStorage := make(map[string]string)

	log.Println("Starting server on ", cfg.ServerAddress)
	err := http.ListenAndServe(cfg.ServerAddress, app.RootRouter(urlStorage, cfg.BaseURL))
	if err != nil {
		panic(err)
	}
}
