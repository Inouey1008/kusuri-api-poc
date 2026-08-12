package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inouey1008/kusuri-api-poc/internal/httpx"
)

func TestWriteJSON(t *testing.T) {
	testCases := []struct {
		name         string
		status       int
		value        any
		expectedBody string
	}{
		{
			name:   `構造体`,
			status: http.StatusOK,
			value: struct {
				Name string `json:"name"`
			}{Name: "エゼチミブ"},
			expectedBody: `{"name":"エゼチミブ"}`,
		},
		{
			name:         `スライス`,
			status:       http.StatusOK,
			value:        []int{1, 2, 3},
			expectedBody: `[1,2,3]`,
		},
		{
			name:         `nil スライスは null になる`,
			status:       http.StatusOK,
			value:        []int(nil),
			expectedBody: `null`,
		},
		{
			name:         `ステータスは引数どおり`,
			status:       http.StatusNotFound,
			value:        map[string]string{"error": "not found"},
			expectedBody: `{"error":"not found"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			httpx.WriteJSON(recorder, testCase.status, testCase.value)

			assert.Equal(t, testCase.status, recorder.Code)
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			assert.JSONEq(t, testCase.expectedBody, recorder.Body.String())
		})
	}
}
