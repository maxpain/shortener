package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maxpain/shortener/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	app := NewApp()
	ts := httptest.NewServer(app.Router)

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
	app := NewApp()
	ts := httptest.NewServer(app.Router)

	originalURL := "https://google.com/"
	response, shortenedURL := testRequest(t, ts, http.MethodPost, "/", strings.NewReader(originalURL))
	defer response.Body.Close()

	assert.NotEmpty(t, shortenedURL)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	hashPath := strings.Replace(string(shortenedURL), *config.BaseURL, "", 1)
	response2, _ := testRequest(t, ts, http.MethodGet, hashPath, nil)
	defer response2.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
	assert.Equal(t, originalURL, response2.Header.Get("Location"))
}

func TestShortenerJSON(t *testing.T) {
	app := NewApp()
	ts := httptest.NewServer(app.Router)

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

	hashPath := strings.Replace(response.Result, *config.BaseURL, "", 1)
	response2, _ := testRequest(t, ts, http.MethodGet, hashPath, nil)
	defer response2.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
	assert.Equal(t, request.URL, response2.Header.Get("Location"))
}
