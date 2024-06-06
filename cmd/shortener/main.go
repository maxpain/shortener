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

func handler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		if req.URL.Path != "/" {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		if req.Header.Get("content-type") != "text/plain" {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(req.Body)

		if err != nil {
			http.Error(res, "Bad Request", http.StatusBadRequest)
			return
		}

		link := string(body)
		hash := generateHash(link)
		links[hash] = link

		res.WriteHeader(http.StatusCreated)
		res.Write([]byte("http://localhost:8080/" + hash))
	case http.MethodGet:
		hash := req.URL.Path[1:]
		link, ok := links[hash]

		if !ok {
			http.Error(res, "Bad Request", http.StatusBadRequest)
		}

		http.Redirect(res, req, link, http.StatusTemporaryRedirect)

	default:
	}
}

func main() {
	err := http.ListenAndServe("localhost:8080", http.HandlerFunc(handler))

	if err != nil {
		panic(err)
	}
}
