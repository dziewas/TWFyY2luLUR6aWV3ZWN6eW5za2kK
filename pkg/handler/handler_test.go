// +build unit !integration

package handler

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		path               string
		payload            string
		limit              int
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "post ok",
			method:             "POST",
			path:               "/hello",
			payload:            "1234567890",
			limit:              100,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "hello",
		},
		{
			name:               "post entity too large",
			method:             "POST",
			path:               "/hello",
			payload:            "1234567890",
			limit:              5,
			expectedStatusCode: http.StatusRequestEntityTooLarge,
			expectedBody:       "\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, "hello")
			}).Methods("POST")

			sizeLimiter := NewSizeLimiter(tc.limit)

			ts := httptest.NewServer(NewChain(router, sizeLimiter))
			client := ts.Client()

			req, err := http.NewRequest(tc.method, ts.URL+tc.path, strings.NewReader(tc.payload))
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}
