package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type want struct {
	code        int
	contentType string
}

type testCase struct {
	name   string
	method string
	want   want
	url    string
	body   string
}

func TestHandler(t *testing.T) {
	testCases := []testCase{
		{
			name:   "POST handler should shorten an url without any errors",
			method: http.MethodPost,
			want: want{
				code:        201,
				contentType: "text/plain",
			},
			url:  "/",
			body: "https://google.com/",
		},
		{
			name:   "POST handler should only respond for /",
			method: http.MethodPost,
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			url:  "/dsadqwdqd",
			body: "",
		},
		{
			name:   "GET handler should return an error if url doesn't exists",
			method: http.MethodGet,
			want: want{
				code:        400,
				contentType: "text/plain",
			},
			url:  "/dsadqwdqd",
			body: "",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.url, strings.NewReader(test.body))
			w := httptest.NewRecorder()
			Handler(w, request)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}

func TestShortener(t *testing.T) {
	originalURL := "https://google.com/"

	request := httptest.NewRequest("POST", "/", strings.NewReader(originalURL))
	w := httptest.NewRecorder()
	Handler(w, request)
	res := w.Result()
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	shortenedURL := string(resBody)
	assert.NotEmpty(t, shortenedURL)

	request = httptest.NewRequest("GET", shortenedURL, nil)
	w = httptest.NewRecorder()
	Handler(w, request)
	res2 := w.Result()
	defer res2.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, res2.StatusCode)
	assert.Equal(t, originalURL, res2.Header.Get("Location"))
}
