package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/logging"
)

// slog の出力先を差し替え、書き込まれた内容を取り出せるようにする
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	buffer := &bytes.Buffer{}
	original := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(buffer, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	return buffer
}

func TestMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		path           string
		handlerStatus  int
		expectedPath   string
		expectedStatus int
	}{
		{
			name:           `200 を記録する`,
			path:           "/drugs?q=エゼチミブ",
			handlerStatus:  http.StatusOK,
			expectedPath:   "/drugs",
			expectedStatus: http.StatusOK,
		},
		{
			name:           `400 を記録する`,
			path:           "/drugs",
			handlerStatus:  http.StatusBadRequest,
			expectedPath:   "/drugs",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			buffer := captureLog(t)

			called := false
			next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				called = true
				writer.WriteHeader(testCase.handlerStatus)
			})

			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			recorder := httptest.NewRecorder()
			logging.Middleware(next).ServeHTTP(recorder, request)

			require.True(t, called, "next が呼ばれていない")
			assert.Equal(t, testCase.handlerStatus, recorder.Code)

			var entry struct {
				Message  string  `json:"msg"`
				Method   string  `json:"method"`
				Path     string  `json:"path"`
				Status   int     `json:"status"`
				Duration float64 `json:"duration"`
			}
			require.NoError(t, json.Unmarshal(buffer.Bytes(), &entry))

			assert.Equal(t, "request", entry.Message)
			assert.Equal(t, http.MethodGet, entry.Method)
			assert.Equal(t, testCase.expectedPath, entry.Path, "クエリ文字列は含めない")
			assert.Equal(t, testCase.expectedStatus, entry.Status)
			assert.Positive(t, entry.Duration)
		})
	}
}

// WriteHeader を呼ばない handler でも 200 として記録される
func TestMiddleware_DefaultStatus(t *testing.T) {
	buffer := captureLog(t)

	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("ok"))
	})

	request := httptest.NewRequest(http.MethodGet, "/drugs", nil)
	logging.Middleware(next).ServeHTTP(httptest.NewRecorder(), request)

	var entry struct {
		Status int `json:"status"`
	}
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &entry))

	assert.Equal(t, http.StatusOK, entry.Status)
}
