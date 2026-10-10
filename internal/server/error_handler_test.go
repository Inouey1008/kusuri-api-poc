package server_test

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/inouey1008/kusuri-api-poc/internal/server"
	"github.com/inouey1008/kusuri-api-poc/internal/validation"
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

// リクエストごとに recordingTransport を積んだ hub を echo.Context に載せるミドルウェア。
// 本番の sentryecho.New の代わりに使い、送信を記録する。
func hubMiddleware(t *testing.T, transport *recordingTransport) echo.MiddlewareFunc {
	t.Helper()

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: transport,
	})
	require.NoError(t, err)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sentryecho.SetHubOnContext(c, sentry.NewHub(client, sentry.NewScope()))
			return next(c)
		}
	}
}

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

	t.Run(`Sentry の hub がなくても panic せず、応答は変わらない`, func(t *testing.T) {
		var recorder *httptest.ResponseRecorder

		require.NotPanics(t, func() {
			recorder = callHandler(t, func(c echo.Context) error {
				return errors.New("something went wrong")
			})
		})

		require.Equal(t, http.StatusInternalServerError, recorder.Code)

		var body errorx.Errorx
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		assert.Equal(t, "INTERNAL_ERROR", body.Code)
	})

	t.Run(`*errorx.Panic は汎用の 500 だけを返し、panic の値を漏らさない`, func(t *testing.T) {
		recorder := callHandler(t, func(c echo.Context) error {
			return errorx.NewPanic("secret: /internal/drug/repository.go")
		})

		require.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.JSONEq(t,
			`{"code":"INTERNAL_ERROR","error":"サーバー内部でエラーが発生しました"}`,
			recorder.Body.String())
		assert.NotContains(t, recorder.Body.String(), "secret")
		assert.NotContains(t, recorder.Body.String(), "repository.go")
	})

	t.Run(`panic は Recovery と errorHandler を通って Sentry へ Fatal で 1 件だけ届き、応答は汎用の 500 になる`, func(t *testing.T) {
		transport := &recordingTransport{}

		e := server.NewWith(
			[]server.Registerer{&handlerRegisterer{handler: func(c echo.Context) error {
				panic("secret: /internal/drug/repository.go")
			}}},
			[]echo.MiddlewareFunc{hubMiddleware(t, transport), middleware.Recovery},
		)

		recorder := sendGet(t, e, "/boom")

		require.Equal(t, http.StatusInternalServerError, recorder.Code)
		assert.NotContains(t, recorder.Body.String(), "secret")

		require.Len(t, transport.events, 1, "送信は errorHandler の 1 件だけ")
		assert.Equal(t, sentry.LevelFatal, transport.events[0].Level)
	})
}
