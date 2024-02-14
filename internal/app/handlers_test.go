package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ma-shulgin/go-link-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootRouter(t *testing.T) {
	InitializeJWT("TEST_TOKEN")
	// "sub" : "test_user"
	token, err := GenerateJWT("test_user")
	require.NoError(t, err)

	filePath := "/tmp/test_db.json"
	os.Remove(filePath)
	store, err := storage.InitFileStore(filePath)
	require.NoError(t, err)
	defer store.Close()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectPing().WillReturnError(nil)

	var originalURLs []string
	for i := 0; i < 4; i++ {
		url := fmt.Sprintf("https://example%d.com", i)
		originalURLs = append(originalURLs, url)
	}


	//err = store.AddURL(context.Background(), originalURL, urlID)
	//require.NoError(t, err)

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
			body:         originalURLs[0],
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/" + GenerateShortURLID(originalURLs[0]),
			responseType: "text",
		},
		{
			name:         "Redirect URL",
			method:       http.MethodGet,
			path:         "/" +  GenerateShortURLID(originalURLs[0]),
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
			method:       http.MethodPut,
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
			body:         `{"url": "` + originalURLs[1] + `"}`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"result": "http://localhost:8080/` +  GenerateShortURLID(originalURLs[1]) + `"}`,
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
			req.AddCookie(&http.Cookie{
				Name:     authCookieName,
				Value:    token,
				HttpOnly: true,
			})

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
