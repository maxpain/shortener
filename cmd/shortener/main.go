package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var links = make(map[string]string)
var length uint8 = 8

func generateHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:length]
}

func GetHandler(rw http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	link, ok := links[hash]

	if !ok {
		http.Error(rw, "Link not found", http.StatusBadRequest)
	}

	http.Redirect(rw, r, link, http.StatusTemporaryRedirect)
}

func PostHandler(rw http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(rw, "Error reading body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	link := string(body)

	if link == "" {
		http.Error(rw, "Empty body", http.StatusBadRequest)
		return
	}

	hash := generateHash(link)
	links[hash] = link

	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte("http://" + r.Host + "/" + hash))
}

func NotFoundHandler(rw http.ResponseWriter, r *http.Request) {
	http.Error(rw, "Not found", http.StatusBadRequest)
}

func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/{hash}", GetHandler)
	r.Post("/", PostHandler)
	r.NotFound(NotFoundHandler)
	r.MethodNotAllowed(NotFoundHandler)

	return r
}

func main() {
	err := http.ListenAndServe("localhost:8080", Router())

	if err != nil {
		panic(err)
	}
}
