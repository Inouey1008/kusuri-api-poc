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

func TestHandler_Check(t *testing.T) {
	handler := server.NewWith([]server.Registerer{health.NewHandler()}, nil)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}
