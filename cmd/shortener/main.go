package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

var links = make(map[string]string)
var length uint8 = 8

func generateHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:length]
}

func Handler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		PostHandler(res, req)
	case http.MethodGet:
		GetHandler(res, req)
	default:
		http.Error(res, "Bad method", http.StatusBadRequest)
	}
}

func GetHandler(res http.ResponseWriter, req *http.Request) {
	hash := req.URL.Path[1:]
	link, ok := links[hash]

	if !ok {
		http.Error(res, "Bad Request", http.StatusBadRequest)
	}

	http.Redirect(res, req, link, http.StatusTemporaryRedirect)
}

func PostHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(res, "Bad path", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(res, "Error reading body", http.StatusBadRequest)
		return
	}

	link := string(body)
	hash := generateHash(link)
	links[hash] = link

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("http://localhost:8080/" + hash))
}

func main() {
	err := http.ListenAndServe("localhost:8080", http.HandlerFunc(Handler))

	if err != nil {
		panic(err)
	}
}
