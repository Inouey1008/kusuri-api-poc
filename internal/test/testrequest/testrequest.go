package testrequest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
	"github.com/inouey1008/kusuri-api-poc/internal/server"
	"github.com/inouey1008/kusuri-api-poc/internal/test/testdb"
)

// 実 DB を使い、本番と同じ構成の API に GET リクエストする
func Get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	cfg := config.Config{
		Environment:  "local",
		Port:         "8080",
		DocsUser:     "docs",
		DocsPassword: "secret",
	}

	handler := server.New(cfg, testdb.Connect(t))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

	return recorder
}
