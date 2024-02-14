package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ma-shulgin/go-link-shortener/internal/logger"
	"github.com/ma-shulgin/go-link-shortener/internal/storage"
	"go.uber.org/zap"
)

func RootRouter(urlStorage storage.URLStore, baseURL string) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.WithLogging)
	r.Use(gzipMiddleware)
	r.Use(JwtAuthMiddleware)

	r.Get("/ping", handlePing(urlStorage))
	r.Get("/{id}", handleRedirect(urlStorage))
	r.Get("/api/user/urls", handleUserUrls(urlStorage, baseURL))
	r.Post("/", handleShorten(urlStorage, baseURL))
	r.Post("/api/shorten", handleAPIShorten(urlStorage, baseURL))
	r.Post("/api/shorten/batch", handleBatchShorten(urlStorage, baseURL))

	return r
}

func handleUserUrls(urlStorage storage.URLStore, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			urls, err := urlStorage.GetUserURLs(ctx)
			if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
			}

			if len(urls) == 0 {
					w.WriteHeader(http.StatusNoContent)
					return
			}
			for i := range urls {
				urls[i].ShortURL = baseURL + "/" + urls[i].ShortURL
		}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(urls)
	}
}

func handlePing(urlStorage storage.URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := urlStorage.Ping(ctx); err != nil {
			http.Error(w, "Database ping failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func handleRedirect(urlStorage storage.URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx:= r.Context()
		urlID := chi.URLParam(r, "id")
		if originalURL, ok := urlStorage.GetURL(ctx,urlID); ok {
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

func handleAPIShorten(urlStorage storage.URLStore, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
		
		w.Header().Set("Content-Type", "application/json")
		err := urlStorage.AddURL(ctx, req.URL, urlID)
		if err != nil {
			if errors.Is(err, storage.ErrConflict) {
				w.WriteHeader(http.StatusConflict)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusCreated)
		}

		resp := shortenResponse{
			Result: baseURL + "/" + urlID,
		}

		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			logger.Log.Debug("error encoding response", zap.Error(err))
			return
		}
		logger.Log.Debug("sending HTTP 201 response")
	}
}

type batchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type batchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func handleBatchShorten(urlStorage storage.URLStore, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx:= r.Context()
		var req []batchRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			logger.Log.Error("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var batchRes []batchResponse
		var urlsToAdd []storage.URLRecord

		for _, req := range req {
			urlID := GenerateShortURLID(req.OriginalURL)
			urlsToAdd = append(urlsToAdd, storage.URLRecord{
				ShortURL:    urlID,
				OriginalURL: req.OriginalURL,
			})
			batchRes = append(batchRes, batchResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      baseURL + "/" + urlID,
			})
		}
		if len(urlsToAdd) == 0 {
			http.Error(w, "Write at least one URL", http.StatusBadRequest)
			return
		}

		if err := urlStorage.AddURLBatch(ctx, urlsToAdd); err != nil {
			logger.Log.Error("Failed to save URLs", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		if err := enc.Encode(batchRes); err != nil {
			logger.Log.Debug("error encoding response", zap.Error(err))
			return
		}
	}
}

func handleShorten(urlStorage storage.URLStore, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		originalURL, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		urlID := GenerateShortURLID(string(originalURL))
		shortenedURL := baseURL + "/" + urlID
		w.Header().Set("Content-Type", "text/plain")

		err = urlStorage.AddURL(ctx, string(originalURL), urlID)
		if err != nil {
			if errors.Is(err, storage.ErrConflict) {
				w.WriteHeader(http.StatusConflict)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		w.Write([]byte(shortenedURL))
	}
}
