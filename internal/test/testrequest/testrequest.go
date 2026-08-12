package testrequest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inouey1008/kusuri-api-poc/internal/server"
	"github.com/inouey1008/kusuri-api-poc/internal/test/testdb"
)

// 実 temp DB を使い、本番と同じ配線で組み上げた API に GET する
func Get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	handler := server.New(testdb.Connect(t))
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

	return recorder
}
