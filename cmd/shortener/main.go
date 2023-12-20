package main

import (
	"log"
	"net/http"

	"github.com/ma-shulgin/go-link-shortener/cmd/config"
	"github.com/ma-shulgin/go-link-shortener/internal/app"
)

func main() {
	cfg := config.GetConfig()
	
	log.Println("Starting server on ", cfg.ServerAddress)
	err := http.ListenAndServe(cfg.ServerAddress, app.RootRouter(cfg.BaseURL))
	if err != nil {
		panic(err)
	}
}
