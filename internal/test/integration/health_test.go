package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/test/testrequest"
)

func TestHealth(t *testing.T) {
	recorder := testrequest.Get(t, "/health")

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}
