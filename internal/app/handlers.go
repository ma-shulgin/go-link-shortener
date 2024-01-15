package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ma-shulgin/go-link-shortener/internal/logger"
	"go.uber.org/zap"
)

func RootRouter(urlStorage map[string]string, baseURL string) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.WithLogging)
	r.Use(gzipMiddleware)

	r.Post("/", handleShorten(urlStorage, baseURL))
	r.Get("/{id}", handleRedirect(urlStorage))
	r.Post("/api/shorten", handleAPIShorten(urlStorage, baseURL))
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
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
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

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

func handleAPIShorten(urlStorage map[string]string, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		logger.Log.Debug("decoding request")
		var req shortenRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			logger.Log.Error("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		urlID := GenerateShortURLID(req.URL)
		urlStorage[urlID] = req.URL

		resp := shortenResponse{
			Result: baseURL + "/" + urlID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			logger.Log.Debug("error encoding response", zap.Error(err))
			return
		}
		logger.Log.Debug("sending HTTP 201 response")
	}
}

