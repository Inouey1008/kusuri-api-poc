package server_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/server"
	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

// 渡されたハンドラを /boom に登録する
type handlerRegisterer struct {
	handler echo.HandlerFunc
}

func (h *handlerRegisterer) Register(instance *echo.Echo) {
	instance.GET("/boom", h.handler)
}

// ハンドラを本番と同じ Echo に載せて呼び出し、返ってきた応答を得る
func callHandler(t *testing.T, handler echo.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	e := server.NewWith([]server.Registerer{&handlerRegisterer{handler: handler}}, nil)

	return sendGet(t, e, "/boom")
}

func TestErrorHandler(t *testing.T) {
	t.Run(`errorx の status と本文をそのまま返す`, func(t *testing.T) {
		details := []validation.FieldError{{Field: "q", Message: "too long"}}

		recorder := callHandler(t, func(c echo.Context) error {
			return errorx.Validation.WithDetails(details)
		})

		require.Equal(t, http.StatusBadRequest, recorder.Code)

		var body errorx.Errorx
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))

		assert.Equal(t, "VALIDATION_FAILED", body.Code)
		assert.Equal(t, details, body.Details)
		assert.Zero(t, body.Status, "Status はボディに含めない")
	})

	t.Run(`Echo の HTTPError もアプリのエラー形式に揃える`, func(t *testing.T) {
		recorder := callHandler(t, func(c echo.Context) error {
			return echo.NewHTTPError(http.StatusMethodNotAllowed)
		})

		require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)

		var body errorx.Errorx
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))

		assert.Equal(t, "METHOD_NOT_ALLOWED", body.Code)
	})

	t.Run(`判別できない error は 500 として返す`, func(t *testing.T) {
		recorder := callHandler(t, func(c echo.Context) error {
			return errors.New("something went wrong")
		})

		require.Equal(t, http.StatusInternalServerError, recorder.Code)

		var body errorx.Errorx
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))

		assert.Equal(t, "INTERNAL_ERROR", body.Code)
		assert.NotContains(t, recorder.Body.String(), "something went wrong",
			"内部のエラー文はクライアントへ返さない")
	})

	t.Run(`応答を書き終えていれば上書きしない`, func(t *testing.T) {
		recorder := callHandler(t, func(c echo.Context) error {
			if err := c.String(http.StatusOK, "partial"); err != nil {
				return err
			}

			return errorx.Internal
		})

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "partial", recorder.Body.String())
	})
}
