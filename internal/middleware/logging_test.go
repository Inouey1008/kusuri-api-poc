package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/middleware"
)

// slog の出力先を差し替え、書き込まれた内容を取り出せるようにする
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	buffer := &bytes.Buffer{}
	original := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(buffer, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	return buffer
}

// path へ GET し、記録された request と response の 2 行を読み出す
func serve(t *testing.T, path string, handler echo.HandlerFunc) (request, response map[string]any) {
	t.Helper()

	buffer := captureLog(t)

	e := echo.New()
	// Logging は c.Error 経由でこれを呼ぶ。本番と同じく status を反映させる
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var target errorx.Errorx
		if errors.As(err, &target) {
			_ = c.JSON(target.Status, target)
		}
	}

	c := e.NewContext(httptest.NewRequest(http.MethodGet, path, nil), httptest.NewRecorder())

	_ = middleware.Logging(handler)(c)

	decoder := json.NewDecoder(buffer)
	require.NoError(t, decoder.Decode(&request))
	require.NoError(t, decoder.Decode(&response))

	return request, response
}

func okHandler(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func TestLogging(t *testing.T) {
	t.Run(`リクエストとレスポンスを同じ request_id で紐付ける`, func(t *testing.T) {
		request, response := serve(t, "/drugs", okHandler)

		assert.Equal(t, "request", request["msg"])
		assert.Equal(t, "response", response["msg"])

		assert.NotEmpty(t, request["request_id"])
		assert.Equal(t, request["request_id"], response["request_id"])
	})

	t.Run(`X-Request-Id を受け取ればそれを使う`, func(t *testing.T) {
		buffer := captureLog(t)

		e := echo.New()
		httpRequest := httptest.NewRequest(http.MethodGet, "/drugs", nil)
		httpRequest.Header.Set(echo.HeaderXRequestID, "given-id")
		c := e.NewContext(httpRequest, httptest.NewRecorder())

		_ = middleware.Logging(okHandler)(c)

		var entry map[string]any
		require.NoError(t, json.NewDecoder(buffer).Decode(&entry))
		assert.Equal(t, "given-id", entry["request_id"], "呼び出し元の追跡 ID を引き継ぐ")
	})

	t.Run(`リクエストにはメソッドとパスとクエリを記録する`, func(t *testing.T) {
		request, _ := serve(t, "/drugs?q=エゼチミブ", okHandler)

		assert.Equal(t, http.MethodGet, request["method"])
		assert.Equal(t, "/drugs", request["path"])
		assert.Equal(t, map[string]any{"q": []any{"エゼチミブ"}}, request["query"])
	})

	t.Run(`クエリ文字列はパスの出力に含まれない`, func(t *testing.T) {
		request, _ := serve(t, "/drugs?q=エゼチミブ", okHandler)

		assert.Equal(t, "/drugs", request["path"], "クエリは query として別に持つ")
	})

	t.Run(`秘密情報にあたるクエリは伏せる`, func(t *testing.T) {
		testCases := []struct {
			name          string
			path          string
			expectedQuery map[string]any
		}{
			{
				name:          `token は伏せる`,
				path:          "/drugs?q=エゼチミブ&token=abc123",
				expectedQuery: map[string]any{"q": []any{"エゼチミブ"}, "token": []any{"***"}},
			},
			{
				name:          `派生形も拾う`,
				path:          "/drugs?access_token=abc123&api_key=xyz",
				expectedQuery: map[string]any{"access_token": []any{"***"}, "api_key": []any{"***"}},
			},
			{
				name:          `keyword は巻き込まない`,
				path:          "/drugs?keyword=エゼチミブ",
				expectedQuery: map[string]any{"keyword": []any{"エゼチミブ"}},
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				request, _ := serve(t, testCase.path, okHandler)

				assert.Equal(t, testCase.expectedQuery, request["query"])
			})
		}
	})

	t.Run(`レスポンスにはステータスと所要時間を記録する`, func(t *testing.T) {
		_, response := serve(t, "/drugs", okHandler)

		assert.Equal(t, float64(http.StatusOK), response["status"])
		assert.Positive(t, response["duration"])
	})

	t.Run(`ハンドラがレスポンスを確定せずに error を返した場合でも、正しい status が出力される`, func(t *testing.T) {
		_, response := serve(t, "/drugs", func(c echo.Context) error {
			return errorx.Validation
		})

		assert.Equal(t, float64(http.StatusBadRequest), response["status"])
	})

	t.Run(`エラーの場合は error の内容が出力される`, func(t *testing.T) {
		_, response := serve(t, "/drugs", func(c echo.Context) error {
			return errorx.Validation
		})

		assert.Equal(t, "入力内容に誤りがあります", response["error"])
	})

	t.Run(`status に応じてレベルを変える`, func(t *testing.T) {
		testCases := []struct {
			name          string
			status        int
			expectedLevel string
		}{
			{name: `2xx は INFO`, status: http.StatusOK, expectedLevel: "INFO"},
			{name: `4xx は WARN`, status: http.StatusBadRequest, expectedLevel: "WARN"},
			{name: `5xx は ERROR`, status: http.StatusInternalServerError, expectedLevel: "ERROR"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				_, response := serve(t, "/drugs", func(c echo.Context) error {
					return c.NoContent(testCase.status)
				})

				assert.Equal(t, testCase.expectedLevel, response["level"])
			})
		}
	})
}

func TestLogging_Body(t *testing.T) {
	post := func(t *testing.T, body string, contentType string, handler echo.HandlerFunc) map[string]any {
		t.Helper()

		buffer := captureLog(t)

		e := echo.New()
		httpRequest := httptest.NewRequest(http.MethodPost, "/drugs", strings.NewReader(body))
		httpRequest.Header.Set(echo.HeaderContentType, contentType)
		c := e.NewContext(httpRequest, httptest.NewRecorder())

		_ = middleware.Logging(handler)(c)

		var entry map[string]any
		require.NoError(t, json.NewDecoder(buffer).Decode(&entry))

		return entry
	}

	t.Run(`JSON ボディを伏せて記録する`, func(t *testing.T) {
		entry := post(t, `{"name":"エゼチミブ","password":"hunter2"}`, echo.MIMEApplicationJSON, okHandler)

		assert.Equal(t, map[string]any{"name": "エゼチミブ", "password": "***"}, entry["body"])
	})

	t.Run(`読んだボディはハンドラでも読み直せる`, func(t *testing.T) {
		body := `{"name":"エゼチミブ"}`

		received := ""
		post(t, body, echo.MIMEApplicationJSON, func(c echo.Context) error {
			raw, err := io.ReadAll(c.Request().Body)
			require.NoError(t, err)
			received = string(raw)

			return c.NoContent(http.StatusOK)
		})

		assert.JSONEq(t, body, received, "ログのために読んでも消費されない")
	})

	t.Run(`JSON でないボディは記録しない`, func(t *testing.T) {
		entry := post(t, "name=エゼチミブ", echo.MIMEApplicationForm, okHandler)

		assert.NotContains(t, entry, "body")
	})
}
