package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/middleware"
)

// recordingTransport は SendEvent で受け取ったイベントを記録するだけの sentry.Transport。
type recordingTransport struct {
	events []*sentry.Event
}

func (t *recordingTransport) Configure(sentry.ClientOptions)        {}
func (t *recordingTransport) SendEvent(event *sentry.Event)         { t.events = append(t.events, event) }
func (t *recordingTransport) Flush(time.Duration) bool              { return true }
func (t *recordingTransport) FlushWithContext(context.Context) bool { return true }
func (t *recordingTransport) Close()                                {}

// newContextWithHub は recordingTransport を積んだ hub を持つ echo.Context を返す。
func newContextWithHub(t *testing.T, transport *recordingTransport) echo.Context {
	t.Helper()

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: transport,
	})
	require.NoError(t, err)

	c := newContext(t)
	sentryecho.SetHubOnContext(c, sentry.NewHub(client, sentry.NewScope()))

	return c
}

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

	t.Run(`panic を *errorx.Panic に変換し、panic の値を保持する`, func(t *testing.T) {
		captureLog(t)
		c := newContext(t)

		err := middleware.Recovery(func(c echo.Context) error {
			panic("something went wrong")
		})(c)

		var target *errorx.Panic
		require.ErrorAs(t, err, &target)
		assert.Equal(t, "something went wrong", target.Value)
		assert.NotEmpty(t, target.StackTrace())
	})

	t.Run(`panic を *errorx.Panic に変換する`, func(t *testing.T) {
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

				var target *errorx.Panic
				require.ErrorAs(t, err, &target)
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

	t.Run(`Sentry へは送信しない (送信は errorHandler に一本化されている)`, func(t *testing.T) {
		captureLog(t)
		transport := &recordingTransport{}
		c := newContextWithHub(t, transport)

		_ = middleware.Recovery(func(c echo.Context) error {
			panic("something went wrong")
		})(c)

		assert.Empty(t, transport.events)
	})

	t.Run(`Sentry の hub がなくても panic しない`, func(t *testing.T) {
		captureLog(t)
		c := newContext(t)

		require.NotPanics(t, func() {
			_ = middleware.Recovery(func(c echo.Context) error {
				panic("something went wrong")
			})(c)
		})
	})
}
