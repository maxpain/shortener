package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/utils"
)

const length = 7

type Storage struct {
	links map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		links: make(map[string]string),
	}
}

func (s *Storage) Save(url string) (string, error) {
	hash := generateHash(url)
	s.links[hash] = url

	fullURL, err := utils.СonstructURL(*config.BaseURL, hash)

	if err != nil {
		return "", errors.New("failed to construct URL")
	}

	return fullURL, nil
}

func (s *Storage) GetURL(hash string) (string, bool) {
	url, found := s.links[hash]

	return url, found
}

func generateHash(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:length]
}
