package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/log"
	"github.com/maxpain/shortener/internal/storage"
)

type Handler struct {
	storage *storage.Storage
	logger  *log.Logger
}

func NewHandler(cfg *config.Configuration, logger *log.Logger) (*Handler, error) {
	storage, err := storage.NewStorage(cfg, logger)

	if err != nil {
		return nil, err
	}

	return &Handler{storage: storage, logger: logger}, nil
}

func (h *Handler) Close() error {
	return h.storage.Close()
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

	shortenedLinks, err := h.storage.Save(r.Context(), []storage.LinkInput{{
		OriginalURL:   url,
		CorrelationID: "",
	}})

	if err != nil || len(shortenedLinks) == 0 {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/plain")

	if shortenedLinks[0].AlreadyExists {
		rw.WriteHeader(http.StatusConflict)
	} else {
		rw.WriteHeader(http.StatusCreated)
	}

	rw.Write([]byte(shortenedLinks[0].ShortURL))
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

	shortenedLinks, err := h.storage.Save(r.Context(), []storage.LinkInput{{
		OriginalURL:   request.URL,
		CorrelationID: "",
	}})

	if err != nil || len(shortenedLinks) == 0 {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Result string `json:"result"`
	}{
		Result: shortenedLinks[0].ShortURL,
	}

	rw.Header().Set("Content-Type", "application/json")

	if shortenedLinks[0].AlreadyExists {
		rw.WriteHeader(http.StatusConflict)
	} else {
		rw.WriteHeader(http.StatusCreated)
	}

	err = json.NewEncoder(rw).Encode(response)

	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ShortenBatchJSON(rw http.ResponseWriter, r *http.Request) {
	var request []struct {
		URL           string `json:"original_url"`
		CorrelationID string `json:"correlation_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(rw, "Error reading body", http.StatusBadRequest)
		return
	}

	var linkInputs []storage.LinkInput
	for _, r := range request {
		linkInputs = append(linkInputs, storage.LinkInput{
			OriginalURL:   r.URL,
			CorrelationID: r.CorrelationID,
		})
	}

	shortenedLinks, err := h.storage.Save(r.Context(), linkInputs)

	if err != nil || len(shortenedLinks) == 0 {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	type ResultLink struct {
		ShortURL      string `json:"short_url"`
		CorrelationID string `json:"correlation_id"`
	}

	var response []ResultLink

	for _, l := range shortenedLinks {
		response = append(response, ResultLink{
			ShortURL:      l.ShortURL,
			CorrelationID: l.CorrelationID,
		})
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
	link, err := h.storage.GetURL(r.Context(), hash)

	if err != nil {
		http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if link == "" {
		http.Error(rw, "Link not found", http.StatusBadRequest)
		return
	}

	http.Redirect(rw, r, link, http.StatusTemporaryRedirect)
}

func (h *Handler) NotFound(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not found", http.StatusBadRequest)
}

func (h *Handler) Ping(rw http.ResponseWriter, r *http.Request) {
	if h.storage.DB == nil {
		http.Error(rw, "Database is not configured", http.StatusInternalServerError)
		return
	}

	err := h.storage.DB.Ping(r.Context())

	if err != nil {
		h.logger.Sugar().Errorf("Database is not available: %v", err.Error())
		http.Error(rw, "Database is not available", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
