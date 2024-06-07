package main

import (
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

func testRequest(t *testing.T, ts *httptest.Server, method, path, body string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(body))
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

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	ts := httptest.NewServer(Router())

	testCases := []testCase{
		{
			name:   "POST handler should shorten an url without any errors",
			method: http.MethodPost,
			code:   201,
			path:   "/",
			body:   "https://google.com/",
		},
		{
			name:   "POST handler should only respond for /",
			method: http.MethodPost,
			code:   400,
			path:   "/dsadqwdqd",
			body:   "",
		},
		{
			name:   "GET handler should return an error if url doesn't exists",
			method: http.MethodGet,
			code:   400,
			path:   "/dsadqwdqd",
			body:   "",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, _ := testRequest(t, ts, test.method, test.path, test.body)
			defer res.Body.Close()

			assert.Equal(t, test.code, res.StatusCode)
		})
	}
}

func TestShortener(t *testing.T) {
	ts := httptest.NewServer(Router())
	originalURL := "https://google.com/"
	response, shortenedURL := testRequest(t, ts, http.MethodPost, "/", originalURL)
	defer response.Body.Close()

	assert.NotEmpty(t, shortenedURL)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	hashPath := strings.Replace(shortenedURL, *config.BaseURL, "", 1)
	response2, _ := testRequest(t, ts, http.MethodGet, hashPath, "")
	defer response2.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, response2.StatusCode)
	assert.Equal(t, originalURL, response2.Header.Get("Location"))
}
