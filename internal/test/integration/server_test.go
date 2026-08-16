package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inouey1008/kusuri-api-poc/internal/test/testrequest"
)

// 未登録のパスは 404 を返す
func TestNotFound(t *testing.T) {
	recorder := testrequest.Get(t, "/unknown")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
