package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
)

const length = 6

type (
	Link struct {
		OriginalURL   string `json:"original_url"`
		CorrelationID string `json:"correlation_id"`
	}

	StoredLink struct {
		*Link
		Hash string `json:"hash"`
	}

	ShortenedLink struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
		Saved         bool   `json:"-"`
	}
)

var ErrNotFound = errors.New("Link not found")

func (l *Link) GetStoredLink() *StoredLink {
	return &StoredLink{
		Link: l,
		Hash: generateHash(l.OriginalURL),
	}
}

func generateHash(url string) string {
	hash := sha256.Sum256([]byte(url))

	return hex.EncodeToString(hash[:])[:length]
}

func (l *StoredLink) GetShortenedLink(baseURL string) (*ShortenedLink, error) {
	url, err := constructURL(baseURL, l.Hash)
	if err != nil {
		return nil, err
	}

	return &ShortenedLink{
		CorrelationID: l.CorrelationID,
		ShortURL:      url,
	}, nil
}

func constructURL(base, postfix string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	relativeURL, err := url.Parse(postfix)
	if err != nil {
		return "", fmt.Errorf("failed to parse postfix: %w", err)
	}

	fullURL := baseURL.ResolveReference(relativeURL)

	return fullURL.String(), nil
}
