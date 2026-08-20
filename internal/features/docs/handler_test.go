package docs_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/features/docs"
	"github.com/inouey1008/kusuri-api-poc/internal/server"
)

func sendGet(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	spec := []byte("openapi: 3.0.0\ninfo:\n  title: test\n")
	e := server.NewWith([]server.Registerer{docs.NewHandler(spec)}, nil)

	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

	return recorder
}

func TestHandler_Spec(t *testing.T) {
	t.Run(`openapi.yaml の中身を返す`, func(t *testing.T) {
		recorder := sendGet(t, "/openapi.yaml")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("Content-Type"), "application/yaml")
		assert.Contains(t, recorder.Body.String(), "openapi: 3.0.0")
	})
}

func TestHandler_Page(t *testing.T) {
	t.Run(`Redoc を表示する HTML を返す`, func(t *testing.T) {
		recorder := sendGet(t, "/docs")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, recorder.Body.String(), `spec-url="/openapi.yaml"`)
	})
}

func TestHandler_Script(t *testing.T) {
	t.Run(`Redoc の JavaScript を返す`, func(t *testing.T) {
		recorder := sendGet(t, "/docs/redoc.standalone.js")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("Content-Type"), "application/javascript")
		assert.NotEmpty(t, recorder.Body.Bytes(), "空なら埋め込みに失敗している")
	})
}
