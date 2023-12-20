package main

import (
	"log"
	"net/http"
	"github.com/ma-shulgin/go-link-shortener/internal/app"
)

func main() {
	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", app.RootRouter())
	if err != nil {
		panic(err)
	}
}
