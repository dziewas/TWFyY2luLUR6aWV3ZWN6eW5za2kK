// +build unit !integration

package handler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler/pkg/store"
	"crawler/pkg/store/memory"
	"crawler/pkg/util"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type httpTestCase struct {
	name               string
	method             string
	path               string
	payload            string
	expectedStatusCode int
	expectedInBody     string
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		httpTestCase
		limit int
	}{
		{
			httpTestCase: httpTestCase{
				name:               "post ok",
				method:             "POST",
				path:               "/hello",
				payload:            "1234567890",
				expectedStatusCode: http.StatusOK,
				expectedInBody:     "hello",
			},
			limit: 100,
		},
		{
			httpTestCase: httpTestCase{
				name:               "post entity too large",
				method:             "POST",
				path:               "/hello",
				payload:            "1234567890",
				expectedStatusCode: http.StatusRequestEntityTooLarge,
			},
			limit: 5,
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

			assert.Contains(t, string(body), tc.expectedInBody)
		})
	}
}

const (
	createValid = `
		{
			"url": "http://localhost:8081/range/1000",
			"interval": 1
		}
	`

	createInvalid = `
		{
			"url": "
		}
	`
)

func TestCreateTask(t *testing.T) {
	tests := []httpTestCase{
		{
			name:               "ok - valid task",
			method:             "POST",
			path:               "/api/fetcher",
			payload:            createValid,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "error - invalid",
			method:             "POST",
			path:               "/api/fetcher",
			payload:            createInvalid,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "error - wrong path",
			method:             "POST",
			path:               "/api/invalid",
			payload:            createInvalid,
			expectedStatusCode: http.StatusNotFound,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := memory.NewMemory()

			resp := makeRequest(t, storage, util.GenID, tc.method, tc.path, tc.payload)
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Contains(t, string(body), tc.expectedInBody)
		})
	}
}

func TestGetTask(t *testing.T) {
	tests := []struct {
		httpTestCase
		idGen func(int64) int64
	}{
		{
			httpTestCase: httpTestCase{
				name:               "ok - valid task",
				method:             "GET",
				path:               "/api/fetcher/123/history",
				expectedStatusCode: http.StatusOK,
				expectedInBody:     "[]",
			},
			idGen: func(_ int64) int64 {
				return 123
			},
		},
		{
			httpTestCase: httpTestCase{
				name:               "error - non existing task",
				method:             "GET",
				path:               "/api/fetcher/321/history",
				expectedStatusCode: http.StatusNotFound,
			},
			idGen: func(_ int64) int64 {
				return 123
			},
		},
		{
			httpTestCase: httpTestCase{
				name:               "error - invalid id",
				method:             "GET",
				path:               "/api/fetcher/invalid123/history",
				expectedStatusCode: http.StatusBadRequest,
			},
			idGen: func(_ int64) int64 {
				return 123
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := memory.NewMemory()

			// create a task first
			resp := makeRequest(t, storage, tc.idGen, "POST", "/api/fetcher", createValid)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			resp = makeRequest(t, storage, tc.idGen, tc.method, tc.path, tc.payload)
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Contains(t, string(body), tc.expectedInBody)
		})
	}
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		httpTestCase
		idGen func(int64) int64
	}{
		{
			httpTestCase: httpTestCase{
				name:               "ok - valid task",
				method:             "DELETE",
				path:               "/api/fetcher/123",
				expectedStatusCode: http.StatusOK,
			},
			idGen: func(_ int64) int64 {
				return 123
			},
		},
		{
			httpTestCase: httpTestCase{
				name:               "error - non existing task",
				method:             "DELETE",
				path:               "/api/fetcher/321",
				expectedStatusCode: http.StatusNotFound,
			},
			idGen: func(_ int64) int64 {
				return 123
			},
		},
		{
			httpTestCase: httpTestCase{
				name:               "error - invalid id",
				method:             "DELETE",
				path:               "/api/fetcher/invalid123",
				expectedStatusCode: http.StatusBadRequest,
			},
			idGen: func(_ int64) int64 {
				return 123
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := memory.NewMemory()

			// create a task first
			resp := makeRequest(t, storage, tc.idGen, "POST", "/api/fetcher", createValid)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			resp = makeRequest(t, storage, tc.idGen, tc.method, tc.path, tc.payload)
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Contains(t, string(body), tc.expectedInBody)
		})
	}
}

func TestGetTasks(t *testing.T) {
	tests := []struct {
		httpTestCase
		createdTasks       int
		expectedTasks      int
		expectedProperties []string
	}{
		{
			httpTestCase: httpTestCase{
				name:               "ok - valid task",
				method:             "GET",
				path:               "/api/fetcher",
				payload:            createValid,
				expectedStatusCode: http.StatusOK,
			},
			createdTasks:       10,
			expectedTasks:      10,
			expectedProperties: []string{"url", "interval", "id"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			storage := memory.NewMemory()

			// create task(s) first
			for i := 0; i < tc.createdTasks; i++ {
				resp := makeRequest(t, storage, util.GenID, "POST", "/api/fetcher", createValid)
				require.Equal(t, http.StatusOK, resp.StatusCode)
			}

			resp := makeRequest(t, storage, util.GenID, tc.method, tc.path, tc.payload)
			assert.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			var tasks []interface{}
			err := json.NewDecoder(resp.Body).Decode(&tasks)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedTasks, len(tasks))
			for _, task := range tasks {
				properties, ok := task.(map[string]interface{})
				require.True(t, ok)

				for _, expectedProperty := range tc.expectedProperties {
					assert.Contains(t, properties, expectedProperty)
				}
			}
		})
	}
}

func makeRequest(t *testing.T, storage store.Store, idGen func(int64) int64, method, path, payload string) *http.Response {
	fetcher := NewFetcher(storage, idGen)
	router := NewRouter(fetcher)

	ts := httptest.NewServer(router)
	client := ts.Client()

	req, err := http.NewRequest(method, ts.URL+path, strings.NewReader(payload))
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}
