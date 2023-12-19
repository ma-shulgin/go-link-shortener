package main

import (
	"log"
	"net/http"

	"github.com/ma-shulgin/go-link-shortener/internal/app"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.HandleRequest)

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
