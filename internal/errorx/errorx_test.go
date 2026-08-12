package errorx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

func TestErrorx_Write(t *testing.T) {
	testCases := []struct {
		name            string
		target          errorx.Errorx
		expectedStatus  int
		expectedCode    string
		expectedDetails int
	}{
		{
			name:           `Internal は 500`,
			target:         errorx.Internal,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
		{
			name:           `Validation は 400`,
			target:         errorx.Validation,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_FAILED",
		},
		{
			name: `Details 付き`,
			target: errorx.Validation.WithDetails([]validation.FieldError{
				{Field: "q", Message: "must be at most 100 characters"},
			}),
			expectedStatus:  http.StatusBadRequest,
			expectedCode:    "VALIDATION_FAILED",
			expectedDetails: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			testCase.target.Write(recorder)

			require.Equal(t, testCase.expectedStatus, recorder.Code)
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			var body struct {
				Status  int    `json:"status"` // json:"-" のため常にゼロ値
				Code    string `json:"code"`
				Message string `json:"error"`
				Details []struct {
					Field   string `json:"field"`
					Message string `json:"message"`
				} `json:"details"`
			}
			require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))

			assert.Equal(t, testCase.expectedCode, body.Code)
			assert.NotEmpty(t, body.Message)
			assert.Zero(t, body.Status, "Status はボディに含めない")
			assert.Len(t, body.Details, testCase.expectedDetails)
		})
	}
}

func TestErrorx_WithDetails(t *testing.T) {
	details := []validation.FieldError{{Field: "q", Message: "too long"}}

	got := errorx.Validation.WithDetails(details)

	assert.Equal(t, details, got.Details)
	assert.Nil(t, errorx.Validation.Details, "元の値を変更してはいけない")
}
