package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/middleware"
)

func newContext(t *testing.T) echo.Context {
	t.Helper()

	e := echo.New()

	return e.NewContext(httptest.NewRequest(http.MethodGet, "/drugs", nil), httptest.NewRecorder())
}

func TestRecovery(t *testing.T) {
	t.Run(`panic しなければ結果をそのまま通す`, func(t *testing.T) {
		buffer := captureLog(t)
		c := newContext(t)

		err := middleware.Recovery(func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
		})(c)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, c.Response().Status)
		assert.Empty(t, buffer.String(), "panic していないのでログを書かない")
	})

	t.Run(`ハンドラが返した error はそのまま通す`, func(t *testing.T) {
		c := newContext(t)

		err := middleware.Recovery(func(c echo.Context) error {
			return errorx.Validation
		})(c)

		var target errorx.Errorx
		require.ErrorAs(t, err, &target)
		assert.Equal(t, "VALIDATION_FAILED", target.Code, "panic 以外は書き換えない")
	})

	t.Run(`panic を Internal に変換する`, func(t *testing.T) {
		testCases := []struct {
			name  string
			panic func()
		}{
			{name: `文字列`, panic: func() { panic("something went wrong") }},
			{name: `error`, panic: func() { panic(http.ErrBodyNotAllowed) }},
			{name: `ランタイムエラー`, panic: func() { var drugs []string; _ = drugs[10] }},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				captureLog(t)
				c := newContext(t)

				var err error
				require.NotPanics(t, func() {
					err = middleware.Recovery(func(c echo.Context) error {
						testCase.panic()
						return nil
					})(c)
				})

				var target errorx.Errorx
				require.ErrorAs(t, err, &target)
				assert.Equal(t, http.StatusInternalServerError, target.Status)
				assert.Equal(t, "INTERNAL_ERROR", target.Code)
			})
		}
	})

	t.Run(`スタックトレース付きで記録する`, func(t *testing.T) {
		buffer := captureLog(t)
		c := newContext(t)

		_ = middleware.Recovery(func(c echo.Context) error {
			panic("something went wrong")
		})(c)

		var entry struct {
			Level   string `json:"level"`
			Message string `json:"msg"`
			Stack   string `json:"stack"`
		}
		require.NoError(t, json.Unmarshal(buffer.Bytes(), &entry))

		assert.Equal(t, "ERROR", entry.Level)
		assert.Equal(t, "panic", entry.Message)
		assert.NotEmpty(t, entry.Stack)
	})
}
