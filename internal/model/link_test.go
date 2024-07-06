package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateHash(t *testing.T) {
	assert := assert.New(t)
	t.Parallel()

	tests := []struct {
		url      string
		expected string
	}{
		{
			url:      "https://google.com",
			expected: "05046f",
		},
		{
			url:      "https://yandex.ru",
			expected: "160009",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()

			assert.Equal(tt.expected, generateHash(tt.url))
		})
	}
}

func TestConstructURL(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	t.Parallel()

	tests := []struct {
		baseURL  string
		postfix  string
		expected string
	}{
		{
			baseURL:  "http://localhost:8080",
			postfix:  "05046f",
			expected: "http://localhost:8080/05046f",
		},
		{
			baseURL:  "https://example.com",
			postfix:  "160009",
			expected: "https://example.com/160009",
		},
	}

	for _, tt := range tests {
		t.Run(tt.baseURL, func(t *testing.T) {
			t.Parallel()

			url, err := constructURL(tt.baseURL, tt.postfix)
			require.NoError(err)
			assert.Equal(tt.expected, url)
		})
	}
}
