package e2e_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inouey1008/kusuri-api-poc/internal/test/testrequest"
)

// feature 横断の振る舞い。未登録のパスは server の mux が 404 を返す
func TestNotFound(t *testing.T) {
	recorder := testrequest.Get(t, "/unknown")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
