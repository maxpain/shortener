package app

import (
	"bytes"
	"encoding/json"
	"io"
	stdlog "log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/maxpain/shortener/config"
	"github.com/maxpain/shortener/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	application *Application
	logger      *log.Logger
	cfg         *config.Configuration
)

func TestMain(m *testing.M) {
	var err error
	logger, err = log.NewLogger()

	if err != nil {
		stdlog.Fatalf("Failed to initialize logger: %v", err)
	}

	cfg = config.NewConfiguration()

	application, err = NewApplication(cfg, logger)

	if err != nil {
		logger.Sugar().Fatalf("Failed to initialize application: %v", err)
	}

	code := m.Run()

	application.Close()
	os.Exit(code)
}

type testCase struct {
	name   string
	method string
	code   int
	path   string
	body   string
}

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, body io.Reader) (*http.Response, []byte) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}

func TestRouter(t *testing.T) {
	ts := httptest.NewServer(application.Router)

	testCases := []testCase{
		{
			name:   "Shorten handler should shorten an url without any errors",
			method: http.MethodPost,
			code:   201,
			path:   "/",
			body:   "https://google.com/",
		},
		{
			name:   "Shorten JSON handler should shorten an url without any errors",
			method: http.MethodPost,
			code:   201,
			path:   "/api/shorten",
			body:   "{\"url\":\"https://google.com/\"}",
		},
		{
			name:   "Shorten JSON handler should shorten a links array without any errors",
			method: http.MethodPost,
			code:   201,
			path:   "/api/shorten/batch",
			body:   "[{\"original_url\":\"https://google.com/\", \"correlation_id\":\"test\"}]",
		},
		{
			name:   "Redirect handler should return an error if an url doesn't exists",
			method: http.MethodGet,
			code:   400,
			path:   "/dsadqwdqd",
			body:   "",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, _ := testRequest(t, ts, test.method, test.path, strings.NewReader(test.body))
			defer res.Body.Close()

			assert.Equal(t, test.code, res.StatusCode)
		})
	}
}

func TestShortener(t *testing.T) {
	ts := httptest.NewServer(application.Router)

	originalURL := "https://google.com/"
	response, shortenedURL := testRequest(t, ts, http.MethodPost, "/", strings.NewReader(originalURL))
	defer response.Body.Close()

	assert.NotEmpty(t, shortenedURL)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	hashPath := strings.Replace(string(shortenedURL), cfg.BaseURL, "", 1)
	response2, _ := testRequest(t, ts, http.MethodGet, hashPath, nil)
	defer response2.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
	assert.Equal(t, originalURL, response2.Header.Get("Location"))
}

func TestShortenerJSON(t *testing.T) {
	ts := httptest.NewServer(application.Router)

	request := struct {
		URL string `json:"url"`
	}{
		URL: "https://google.com/",
	}

	requestJSON, err := json.Marshal(request)
	require.NoError(t, err)

	r, body := testRequest(t, ts, http.MethodPost, "/api/shorten", bytes.NewReader(requestJSON))
	defer r.Body.Close()

	assert.NotEmpty(t, body)
	assert.Equal(t, http.StatusCreated, r.StatusCode)

	var response struct {
		Result string `json:"result"`
	}

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.Result)

	hashPath := strings.Replace(response.Result, cfg.BaseURL, "", 1)
	response2, _ := testRequest(t, ts, http.MethodGet, hashPath, nil)
	defer response2.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
	assert.Equal(t, request.URL, response2.Header.Get("Location"))
}

func TestShortenerBatchJSON(t *testing.T) {
	ts := httptest.NewServer(application.Router)

	request := []struct {
		OriginalURL   string `json:"original_url"`
		CorrelationID string `json:"correlation_id"`
	}{{
		OriginalURL:   "https://google.com/",
		CorrelationID: "test",
	}}

	requestJSON, err := json.Marshal(request)
	require.NoError(t, err)

	r, body := testRequest(t, ts, http.MethodPost, "/api/shorten/batch", bytes.NewReader(requestJSON))
	defer r.Body.Close()

	assert.NotEmpty(t, body)
	assert.Equal(t, http.StatusCreated, r.StatusCode)

	var response []struct {
		ShortURL      string `json:"short_url"`
		CorrelationID string `json:"correlation_id"`
	}

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response)
	assert.NotEmpty(t, response[0].ShortURL)
	assert.Equal(t, request[0].CorrelationID, response[0].CorrelationID)

	hashPath := strings.Replace(response[0].ShortURL, cfg.BaseURL, "", 1)
	response2, _ := testRequest(t, ts, http.MethodGet, hashPath, nil)
	defer response2.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
	assert.Equal(t, request[0].OriginalURL, response2.Header.Get("Location"))
}
