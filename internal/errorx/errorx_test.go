package errorx_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

func TestErrorx_Error(t *testing.T) {
	t.Run(`error として扱える`, func(t *testing.T) {
		var err error = errorx.Validation

		assert.EqualError(t, err, "入力内容に誤りがあります")
	})

	t.Run(`errors.As で元の型に戻せる`, func(t *testing.T) {
		var err error = errorx.Validation

		var target errorx.Errorx
		ok := errors.As(err, &target)

		assert.True(t, ok, "HTTPErrorHandler がこの判定で status を決める")
		assert.Equal(t, http.StatusBadRequest, target.Status)
		assert.Equal(t, "VALIDATION_FAILED", target.Code)
	})
}

func TestErrorx_WithDetails(t *testing.T) {
	t.Run(`Details が設定した値で返る`, func(t *testing.T) {
		details := []validation.FieldError{{Field: "q", Message: "too long"}}

		got := errorx.Validation.WithDetails(details)

		assert.Equal(t, details, got.Details)
	})
}
