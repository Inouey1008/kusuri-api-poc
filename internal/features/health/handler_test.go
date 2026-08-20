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

func callCheck(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()

	e := server.NewWith([]server.Registerer{health.NewHandler()}, nil)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	return recorder
}

func TestHandler_Check(t *testing.T) {
	t.Run(`200 を返す`, func(t *testing.T) {
		recorder := callCheck(t)

		require.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run(`status ok を JSON で返す`, func(t *testing.T) {
		recorder := callCheck(t)

		assert.Contains(t, recorder.Header().Get("Content-Type"), "application/json")
		assert.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
	})
}
