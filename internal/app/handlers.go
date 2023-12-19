package app

import (
	"io"
	"net/http"
)

var ShortURLs = make(map[string]string)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		HandleShorten(w, r)
	case http.MethodGet:
		HandleRedirect(w, r)
	default:
		http.Error(w, "Unsupported HTTP method", http.StatusBadRequest)
	}
}

func HandleShorten(w http.ResponseWriter, r *http.Request) {

	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	r.Body.Close()

	urlID := GenerateShortURLID(string(originalURL))
	ShortURLs[urlID] = string(originalURL)

	shortenedURL := "http://localhost:8080/" + urlID
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(shortenedURL))
}

func HandleRedirect(w http.ResponseWriter, r *http.Request) {

	urlID := r.URL.Path[1:]
	if originalURL, ok := ShortURLs[urlID]; ok {
		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
		return
	}

	http.Error(w, "Bad request", http.StatusBadRequest)
}
