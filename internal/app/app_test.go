package app

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/maxpain/shortener/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initApp() (*App, error) {
	ctx := context.Background()
	cfg := config.New(
		config.WithFileStoragePath(""),
	)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	return New(ctx, cfg, logger)
}

func TestRouter(t *testing.T) {
	t.Parallel()

	app, err := initApp()
	require.NoError(t, err)
	t.Cleanup(app.Close)

	tests := []struct {
		name          string
		method        string
		body          string
		path          string
		statusCode    int
		response      string
		isJSON        bool
		checkResponse bool
		location      string
		checkLocation bool
	}{
		{
			name:       "ping",
			method:     "GET",
			path:       "/ping",
			statusCode: fiber.StatusOK,
		},
		{
			name:       "Non-existent redirect",
			method:     "GET",
			path:       "/non-existent",
			statusCode: fiber.StatusNotFound,
		},
		{
			name:          "Shorten single plain",
			method:        "POST",
			path:          "/",
			statusCode:    fiber.StatusCreated,
			body:          "https://google.com",
			response:      "http://localhost:8080/05046f",
			checkResponse: true,
		},
		{
			name:          "Try to shorten single plain second time",
			method:        "POST",
			path:          "/",
			statusCode:    fiber.StatusConflict,
			body:          "https://google.com",
			response:      "http://localhost:8080/05046f",
			checkResponse: true,
		},
		{
			name:          "Shorten single JSON",
			method:        "POST",
			path:          "/api/shorten",
			statusCode:    fiber.StatusCreated,
			body:          `{"url":"https://yandex.ru"}`,
			response:      `{"result":"http://localhost:8080/160009"}`,
			checkResponse: true,
			isJSON:        true,
		},
		{
			name:          "Shorten batch JSON",
			method:        "POST",
			path:          "/api/shorten/batch",
			statusCode:    fiber.StatusCreated,
			body:          `[{"original_url":"https://x.com/","correlation_id":"x"},{"original_url":"https://t.me/","correlation_id":"t"}]`,
			response:      `[{"short_url":"http://localhost:8080/326a64","correlation_id":"x"},{"short_url":"http://localhost:8080/e70e7a","correlation_id":"t"}]`,
			checkResponse: true,
			isJSON:        true,
		},
		{
			name:          "Redirect",
			method:        "GET",
			path:          "/160009",
			statusCode:    fiber.StatusTemporaryRedirect,
			location:      "https://yandex.ru",
			checkLocation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))

			if tt.isJSON {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := app.Test(req)
			require.NoError(err)

			assert.Equal(tt.statusCode, resp.StatusCode, "status code")

			if tt.checkResponse {
				body, err := io.ReadAll(resp.Body)
				require.NoError(err)

				if tt.isJSON {
					assert.JSONEq(tt.response, string(body), "response json")
				} else {
					assert.Equal(tt.response, string(body), "response body")
				}
			}

			if tt.checkLocation {
				location := resp.Header.Get("Location")
				assert.Equal(tt.location, location, "Location header")
			}

			resp.Body.Close()
		})
	}
}
