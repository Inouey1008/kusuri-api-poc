package logging_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/logging"
)

// log の出力先を差し替え、書き込まれた内容を取り出せるようにする
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	buffer := &bytes.Buffer{}
	flags := log.Flags()

	log.SetOutput(buffer)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(flags)
	})

	return buffer
}

func TestMiddleware(t *testing.T) {
	buffer := captureLog(t)

	called := false
	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		called = true
		writer.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodGet, "/drugs?q=エゼチミブ", nil)
	recorder := httptest.NewRecorder()
	logging.Middleware(next).ServeHTTP(recorder, request)

	require.True(t, called, "next が呼ばれていない")
	assert.Equal(t, http.StatusOK, recorder.Code)

	output := buffer.String()
	assert.Contains(t, output, http.MethodGet)
	assert.Contains(t, output, "/drugs") // クエリ文字列は含めない
	assert.NotContains(t, output, "q=エゼチミブ")
}
