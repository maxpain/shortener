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
			assert.Equal(t, test.want.code, res.StatusCode)
		})
	}
}

func TestShortener(t *testing.T) {
	originalUrl := "https://google.com/"

	request := httptest.NewRequest("POST", "/", strings.NewReader(originalUrl))
	w := httptest.NewRecorder()
	Handler(w, request)
	res := w.Result()

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	shortenedUrl := string(resBody)
	assert.NotEmpty(t, shortenedUrl)

	request = httptest.NewRequest("GET", shortenedUrl, nil)
	w = httptest.NewRecorder()
	Handler(w, request)
	res = w.Result()

	assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
	assert.Equal(t, originalUrl, res.Header.Get("Location"))
}
