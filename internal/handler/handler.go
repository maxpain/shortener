package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/maxpain/shortener/internal/storage"
)

type Handler struct {
	storage *storage.Storage
}

func NewHandler(storage *storage.Storage) *Handler {
	return &Handler{storage: storage}
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
