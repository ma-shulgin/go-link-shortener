package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleShortenAndRedirect(t *testing.T) {
	// Setup
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
			expectedCode: http.StatusBadRequest,
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
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			w := httptest.NewRecorder()

			HandleRequest(w, req)

			assert.Equal(t, tc.expectedCode, w.Code, "Response status code does not match expected")
			if tc.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tc.expectedBody, "Response body does not match expected")
			}
		})
	}
}
