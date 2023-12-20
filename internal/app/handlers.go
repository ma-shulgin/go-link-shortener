package app

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RootRouter(urlStorage map[string]string, baseURL string) chi.Router {
	r := chi.NewRouter()

	r.Post("/", handleShorten(urlStorage, baseURL))
	r.Get("/{id}", handleRedirect(urlStorage))

	return r
}

func handleShorten(urlStorage map[string]string, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		originalURL, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		urlID := GenerateShortURLID(string(originalURL))
		urlStorage[urlID] = string(originalURL)

		shortenedURL := baseURL + "/" + urlID
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(shortenedURL))
	}
}

func handleRedirect(urlStorage map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlID := chi.URLParam(r, "id")
		if originalURL, ok := urlStorage[urlID]; ok {
			http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
			return
		}
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}
