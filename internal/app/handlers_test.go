package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootRouter(t *testing.T) {

	originalURL := "https://example.com"
	urlID := GenerateShortURLID(originalURL)
	ShortURLs[urlID] = originalURL

	testCases := []struct {
		name         string
		method       string
		path         string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Shorten URL",
			method:       http.MethodPost,
			path:         "/",
			body:         originalURL,
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/" + urlID,
		},
		{
			name:         "Redirect URL",
			method:       http.MethodGet,
			path:         "/" + urlID,
			expectedCode: http.StatusTemporaryRedirect,
			expectedBody: "",
		},
		{
			name:         "Wrong URL path",
			method:       http.MethodGet,
			path:         "/other/path",
			expectedCode: http.StatusNotFound,
			expectedBody: "",
		},
		{
			name:         "Wrong URL",
			method:       http.MethodGet,
			path:         "/ntexst66",
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "Unsupported Method",
			method:       http.MethodDelete,
			path:         "/",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
	}

	ts := httptest.NewServer(RootRouter("http://localhost:8080/"))
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer ts.Close()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			var err error
			url := ts.URL + tc.path

			if tc.body != "" {
				req, err = http.NewRequest(tc.method, url, bytes.NewBufferString(tc.body))
			} else {
				req, err = http.NewRequest(tc.method, url, nil)
			}
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode, "Response status code does not match expected")

			if tc.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Contains(t, string(body), tc.expectedBody, "Response body does not match expected")
			}
		})
	}
}
