package health_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/features/health"
	"github.com/inouey1008/kusuri-api-poc/internal/server"
)

func get(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()

	handler := server.NewWith([]server.Registerer{health.NewHandler()}, nil)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	return recorder
}

func TestHandler_Check(t *testing.T) {
	t.Run(`200 を返す`, func(t *testing.T) {
		recorder := get(t)

		require.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run(`status ok を JSON で返す`, func(t *testing.T) {
		recorder := get(t)

		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
	})
}
