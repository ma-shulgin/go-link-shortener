package app

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ma-shulgin/go-link-shortener/internal/storage"
)

func TestRootRouter(t *testing.T) {

	
	store, err := storage.InitFileStore("/tmp/test_db.json")
	require.NoError(t, err)
	defer store.Close() 

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
    if err != nil {
        t.Fatal(err)
    }
  defer db.Close()
	mock.ExpectPing().WillReturnError(nil)

	originalURL := "https://example.com"
	urlID := GenerateShortURLID(originalURL)
	err = store.AddURL(originalURL, urlID)
	require.NoError(t, err)

	ts := httptest.NewServer(RootRouter(store, "http://localhost:8080"))

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer ts.Close()

	testCases := []struct {
		name         string
		method       string
		path         string
		body         string
		expectedCode int
		expectedBody string
		responseType string
	}{
		{
			name:         "Ping",
			method:       http.MethodGet,
			path:         "/ping",
			expectedCode: http.StatusOK,
			expectedBody: "OK",
			responseType: "text",
		},
		{
			name:         "Shorten URL",
			method:       http.MethodPost,
			path:         "/",
			body:         originalURL,
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/" + urlID,
			responseType: "text",
		},
		{
			name:         "Redirect URL",
			method:       http.MethodGet,
			path:         "/" + urlID,
			expectedCode: http.StatusTemporaryRedirect,
			expectedBody: "",
			responseType: "",
		},
		{
			name:         "Wrong URL path",
			method:       http.MethodGet,
			path:         "/other/path",
			expectedCode: http.StatusNotFound,
			expectedBody: "",
			responseType: "",
		},
		{
			name:         "Wrong URL",
			method:       http.MethodGet,
			path:         "/ntexst66",
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
			responseType: "",
		},
		{
			name:         "Unsupported Method",
			method:       http.MethodDelete,
			path:         "/",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "API Shorten POST without body",
			method:       http.MethodPost,
			path:         "/api/shorten",
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
			responseType: "",
		},
		{
			name:         "API Shorten POST with body",
			method:       http.MethodPost,
			path:         "/api/shorten",
			body:         `{"url": "https://example.com"}`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"result": "http://localhost:8080/` + urlID + `"}`,
			responseType: "json",
		},
		{
			name:         "API Shorten with Unsupported Method",
			method:       http.MethodGet,
			path:         "/api/shorten",
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
			responseType: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			var err error
			url := ts.URL + tc.path

			if tc.body != "" {
				req, err = http.NewRequest(tc.method, url, bytes.NewBufferString(tc.body))
				req.Header.Set("Content-Type", "application/json")
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
				switch tc.responseType {
				case "json":
					assert.JSONEq(t, tc.expectedBody, string(body), "Response body didn't match expected")

				case "text":
					assert.Equal(t, tc.expectedBody, string(body), "Response body didn't match expected")
				}
			}
		})
	}
}
