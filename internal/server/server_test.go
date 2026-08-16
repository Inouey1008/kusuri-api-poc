package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/server"
)

type stubRegisterer struct {
	path string
	body string
}

func (s *stubRegisterer) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET "+s.path, func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(s.body))
	})
}

func recordingMiddleware(name string, order *[]string) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			*order = append(*order, name)
			next.ServeHTTP(writer, request)
		})
	}
}

func TestNewWith_Routing(t *testing.T) {
	handler := server.NewWith([]server.Registerer{
		&stubRegisterer{path: "/drugs", body: "drugs"},
		&stubRegisterer{path: "/facilities", body: "facilities"},
	}, nil)

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           `1 つ目の Registerer のルート`,
			path:           "/drugs",
			expectedStatus: http.StatusOK,
			expectedBody:   "drugs",
		},
		{
			name:           `2 つ目の Registerer のルート`,
			path:           "/facilities",
			expectedStatus: http.StatusOK,
			expectedBody:   "facilities",
		},
		{
			name:           `未登録のパスは 404`,
			path:           "/unknown",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)

			require.Equal(t, testCase.expectedStatus, recorder.Code)
			if testCase.expectedBody != "" {
				assert.Equal(t, testCase.expectedBody, recorder.Body.String())
			}
		})
	}
}

func TestNewWith_MiddlewareOrder(t *testing.T) {
	order := []string{}
	handler := server.NewWith(
		[]server.Registerer{&stubRegisterer{path: "/drugs", body: "drugs"}},
		[]server.Middleware{
			recordingMiddleware("first", &order),
			recordingMiddleware("second", &order),
		},
	)

	request := httptest.NewRequest(http.MethodGet, "/drugs", nil)
	handler.ServeHTTP(httptest.NewRecorder(), request)

	assert.Equal(t, []string{"first", "second"}, order, "先に定義した middleware から実行される")
}
