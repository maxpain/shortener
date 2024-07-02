package storage

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"

	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/log"
	"github.com/maxpain/shortener/internal/utils"
	"go.uber.org/zap"
)

const length = 7

type Storage struct {
	links   map[string]string
	file    *os.File
	baseURL string
}

type linkData struct {
	Hash string `json:"hash"`
	URL  string `json:"url"`
}

func NewStorage(cfg *config.Configuration, logger *log.Logger) (*Storage, error) {
	s := &Storage{
		links:   make(map[string]string),
		file:    nil,
		baseURL: cfg.BaseURL,
	}

	if cfg.FileStoragePath != "" {
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)

		if err != nil {
			return nil, err
		}

		s.file = file

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()

			var link linkData

			err := json.Unmarshal([]byte(line), &link)

			if err != nil {
				return nil, err
			}

			logger.Debug("Loaded link",
				zap.String("hash", link.Hash),
				zap.String("url", link.URL),
			)

			s.links[link.Hash] = link.URL
		}
	}

	return s, nil
}

func (s *Storage) Save(url string) (string, error) {
	hash := generateHash(url)

	if _, ok := s.links[hash]; !ok && s.file != nil {
		link := linkData{
			URL:  url,
			Hash: hash,
		}

		jsonString, err := json.Marshal(link)

		if err != nil {
			return "", err
		}

		_, err = s.file.WriteString(string(jsonString) + "\n")

		if err != nil {
			return "", err
		}
	}

	s.links[hash] = url
	fullURL, err := utils.СonstructURL(s.baseURL, hash)

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
