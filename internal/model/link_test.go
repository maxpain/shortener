package model_test

import (
	"testing"

	"github.com/maxpain/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStoredLink(t *testing.T) {
	assert := assert.New(t)
	t.Parallel()

	tests := []struct {
		originalURL  string
		expectedHash string
	}{
		{
			originalURL:  "https://google.com",
			expectedHash: "05046f",
		},
		{
			originalURL:  "https://yandex.ru",
			expectedHash: "160009",
		},
	}

	for _, tt := range tests {
		t.Run(tt.originalURL, func(t *testing.T) {
			t.Parallel()

			link := &model.Link{OriginalURL: tt.originalURL}
			storedLink := link.GetStoredLink("test-user-id")

			assert.Equal(tt.expectedHash, storedLink.Hash)
			assert.Equal(tt.originalURL, storedLink.OriginalURL)
			assert.Equal("test-user-id", storedLink.UserID)
		})
	}
}

func TestGetShortenedLink(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	t.Parallel()

	tests := []struct {
		originalURL  string
		baseURL      string
		expectedHash string
		expectedURL  string
	}{
		{
			originalURL:  "https://google.com",
			baseURL:      "http://localhost:8080",
			expectedHash: "05046f",
			expectedURL:  "http://localhost:8080/05046f",
		},
		{
			originalURL:  "https://yandex.ru",
			baseURL:      "https://example.com",
			expectedHash: "160009",
			expectedURL:  "https://example.com/160009",
		},
	}

	for _, tt := range tests {
		t.Run(tt.originalURL, func(t *testing.T) {
			t.Parallel()

			link := &model.Link{
				OriginalURL:   tt.originalURL,
				CorrelationID: "test-correlation-id",
			}

			storedLink := link.GetStoredLink("test-user-id")
			shortenedLink, err := storedLink.GetShortenedLink(tt.baseURL)

			require.NoError(err)
			assert.Equal(tt.expectedURL, shortenedLink.ShortURL)
			assert.Equal("test-correlation-id", shortenedLink.CorrelationID)
		})
	}
}
