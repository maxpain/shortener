package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/maxpain/shortener/internal/db"
	"github.com/maxpain/shortener/internal/log"
	"github.com/maxpain/shortener/internal/storage"
)

type Handler struct {
	storage *storage.Storage
	logger  *log.Logger
	DB      *db.Database
}

func NewHandler(storage *storage.Storage, logger *log.Logger, DB *db.Database) *Handler {
	return &Handler{storage: storage, logger: logger, DB: DB}
}

func (h *Handler) Shorten(rw http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(rw, "Error reading body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	url := string(body)

	if url == "" {
		http.Error(rw, "Empty body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.storage.Save(url)

	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/plain")
	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(shortURL))
}

func (h *Handler) ShortenJSON(rw http.ResponseWriter, r *http.Request) {
	var request struct {
		URL string `json:"url"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(rw, "Error reading body", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(rw, "Empty body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.storage.Save(request.URL)

	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
	}

	response := struct {
		Result string `json:"result"`
	}{
		Result: shortURL,
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(rw).Encode(response)

	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Redirect(rw http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	link, ok := h.storage.GetURL(hash)

	if !ok {
		http.Error(rw, "Link not found", http.StatusBadRequest)
	}

	http.Redirect(rw, r, link, http.StatusTemporaryRedirect)
}

func (h *Handler) NotFound(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not found", http.StatusBadRequest)
}

func (h *Handler) Ping(rw http.ResponseWriter, r *http.Request) {
	err := h.DB.Ping(context.Background())

	if err != nil {
		h.logger.Sugar().Errorf("Database is not available: %v", err.Error())
		http.Error(rw, "Database is not available", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
