package sentryx_test

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/sentryx"
)

// mockTransport は SendEvent で受け取ったイベントを記録するだけの sentry.Transport。
// 実際の送信は行わない。
type mockTransport struct {
	events []*sentry.Event
}

func (t *mockTransport) Configure(sentry.ClientOptions)        {}
func (t *mockTransport) SendEvent(event *sentry.Event)         { t.events = append(t.events, event) }
func (t *mockTransport) Flush(time.Duration) bool              { return true }
func (t *mockTransport) FlushWithContext(context.Context) bool { return true }
func (t *mockTransport) Close()                                {}

// newContextWithHub は mockTransport を積んだ hub を持つ echo.Context を返す。
// sentryecho.GetHubFromContext が拾えるよう、echo 側の SetHubOnContext で載せる。
func newContextWithHub(t *testing.T, transport *mockTransport) echo.Context {
	t.Helper()

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://public@example.com/1",
		Transport: transport,
	})
	require.NoError(t, err)

	hub := sentry.NewHub(client, sentry.NewScope())

	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/drugs", nil), httptest.NewRecorder())
	sentryecho.SetHubOnContext(c, hub)

	return c
}

// withRestoredHub は sentry.CurrentHub() (グローバルな hub) の client を退避し、
// テスト終了後に元へ戻す。Init はグローバル状態を書き換えるため、他のテストに
// 影響しないよう必ず併用する。
func withRestoredHub(t *testing.T) {
	t.Helper()

	hub := sentry.CurrentHub()
	original := hub.Client()
	t.Cleanup(func() { hub.BindClient(original) })
}

// captureLog は slog の出力先を差し替え、書き込まれた内容を取り出せるようにする。
// グローバルな default logger を触るため、終了時に必ず元へ戻す。
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	buffer := &bytes.Buffer{}
	original := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(buffer, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	return buffer
}

func TestInit(t *testing.T) {
	t.Run(`local 環境で DSN が空なら nil を返し、Info ログを出して client を差し替えない`, func(t *testing.T) {
		withRestoredHub(t)
		logs := captureLog(t)
		before := sentry.CurrentHub().Client()

		var err error
		require.NotPanics(t, func() {
			err = sentryx.Init(config.Config{Environment: "local"})
		})

		require.NoError(t, err)
		assert.Same(t, before, sentry.CurrentHub().Client(), "DSN 未設定なら client を差し替えない")
		assert.Contains(t, logs.String(), `"level":"INFO"`)
		assert.NotContains(t, logs.String(), `"level":"WARN"`)
	})

	for _, env := range []string{"dev", "prod"} {
		t.Run(env+` 環境で DSN が空なら nil を返し、Warn ログを出して client を差し替えない`, func(t *testing.T) {
			withRestoredHub(t)
			logs := captureLog(t)
			before := sentry.CurrentHub().Client()

			var err error
			require.NotPanics(t, func() {
				err = sentryx.Init(config.Config{Environment: env})
			})

			require.NoError(t, err)
			assert.Same(t, before, sentry.CurrentHub().Client(), "空 DSN では sentry.Init を呼ばない")
			assert.Contains(t, logs.String(), `"level":"WARN"`)
		})
	}

	t.Run(`dev 環境で DSN を設定すると nil を返し、Environment・TracesSampleRate が client に反映される`, func(t *testing.T) {
		withRestoredHub(t)

		cfg := config.Config{
			SentryDSN:              "https://public@example.com/1",
			Environment:            "dev",
			SentryTracesSampleRate: 0.5,
		}

		err := sentryx.Init(cfg)
		require.NoError(t, err)

		client := sentry.CurrentHub().Client()
		require.NotNil(t, client)

		options := client.Options()
		assert.Equal(t, "dev", options.Environment)
		assert.Equal(t, 0.5, options.TracesSampleRate)
	})

	t.Run(`不正な DSN なら error を返す`, func(t *testing.T) {
		withRestoredHub(t)

		cfg := config.Config{Environment: "dev", SentryDSN: "not-a-valid-dsn"}

		var err error
		require.NotPanics(t, func() {
			err = sentryx.Init(cfg)
		})

		assert.Error(t, err)
	})
}

func TestReport(t *testing.T) {
	t.Run(`hub がなければ panic せず何もしない`, func(t *testing.T) {
		e := echo.New()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/drugs", nil), httptest.NewRecorder())

		require.NotPanics(t, func() {
			sentryx.Report(c, assert.AnError)
		})
	})

	t.Run(`通常の error は Error レベルで 1 件送信する`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)

		sentryx.Report(c, assert.AnError)

		require.Len(t, transport.events, 1)
		assert.Equal(t, assert.AnError.Error(), transport.events[0].Exception[0].Value)
		assert.Equal(t, sentry.LevelError, transport.events[0].Level)
	})

	t.Run(`*errorx.Panic は Fatal レベルで 1 件送信し、スタックトレースが付く`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)

		sentryx.Report(c, errorx.NewPanic("something went wrong"))

		require.Len(t, transport.events, 1)
		event := transport.events[0]
		assert.Equal(t, sentry.LevelFatal, event.Level)

		require.NotEmpty(t, event.Exception)
		stacktrace := event.Exception[len(event.Exception)-1].Stacktrace
		require.NotNil(t, stacktrace, "Panic のスタックが Sentry のイベントへ載っていない")
		assert.NotEmpty(t, stacktrace.Frames)
	})

	t.Run(`*errorx.Panic に包まれていても Fatal レベルで送信する`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)

		sentryx.Report(c, fmt.Errorf("wrapped: %w", errorx.NewPanic("something went wrong")))

		require.Len(t, transport.events, 1)
		assert.Equal(t, sentry.LevelFatal, transport.events[0].Level)
	})

	t.Run(`Panic の後に同じ hub で送る通常の error へ Fatal が漏れない`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)

		sentryx.Report(c, errorx.NewPanic("something went wrong"))
		sentryx.Report(c, assert.AnError)

		require.Len(t, transport.events, 2)
		assert.Equal(t, sentry.LevelFatal, transport.events[0].Level)
		assert.Equal(t, sentry.LevelError, transport.events[1].Level, "WithScope により Fatal は 1 件目に閉じる")
	})

	t.Run(`echo 由来の 4xx は送信しない`, func(t *testing.T) {
		testCases := []struct {
			name string
			err  error
		}{
			{name: `404`, err: echo.ErrNotFound},
			{name: `405`, err: echo.ErrMethodNotAllowed},
			{name: `401`, err: echo.ErrUnauthorized},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				transport := &mockTransport{}
				c := newContextWithHub(t, transport)

				sentryx.Report(c, testCase.err)

				assert.Empty(t, transport.events)
			})
		}
	})

	t.Run(`echo 由来の 5xx は送信する`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)

		sentryx.Report(c, echo.NewHTTPError(http.StatusInternalServerError))

		assert.Len(t, transport.events, 1)
	})

	t.Run(`アプリの 4xx (errorx.Validation) は送信する`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)

		sentryx.Report(c, errorx.Validation)

		assert.Len(t, transport.events, 1)
	})

	t.Run(`レスポンスヘッダの X-Request-Id を request_id タグに載せる`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)
		c.Response().Header().Set(echo.HeaderXRequestID, "abc")

		sentryx.Report(c, assert.AnError)

		require.Len(t, transport.events, 1)
		assert.Equal(t, "abc", transport.events[0].Tags["request_id"])
	})

	t.Run(`X-Request-Id がなければ request_id タグは付かない`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)

		sentryx.Report(c, assert.AnError)

		require.Len(t, transport.events, 1)
		assert.NotContains(t, transport.events[0].Tags, "request_id")
	})

	t.Run(`200 文字を超える X-Request-Id は 200 文字に切り詰める`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)
		c.Response().Header().Set(echo.HeaderXRequestID, strings.Repeat("a", 300))

		sentryx.Report(c, assert.AnError)

		require.Len(t, transport.events, 1)
		assert.Equal(t, strings.Repeat("a", 200), transport.events[0].Tags["request_id"])
	})

	t.Run(`request_id タグは同じ hub の次の送信へ漏れない`, func(t *testing.T) {
		transport := &mockTransport{}
		c := newContextWithHub(t, transport)

		c.Response().Header().Set(echo.HeaderXRequestID, "abc")
		sentryx.Report(c, assert.AnError)
		c.Response().Header().Del(echo.HeaderXRequestID)
		sentryx.Report(c, assert.AnError)

		require.Len(t, transport.events, 2)
		assert.Equal(t, "abc", transport.events[0].Tags["request_id"])
		assert.NotContains(t, transport.events[1].Tags, "request_id", "WithScope によりタグは 1 件目に閉じる")
	})
}

func TestMaskEvent(t *testing.T) {
	t.Run(`Request が nil でも panic せずそのまま返す`, func(t *testing.T) {
		event := &sentry.Event{}

		var got *sentry.Event
		require.NotPanics(t, func() {
			got = sentryx.MaskEvent(event, nil)
		})

		assert.Same(t, event, got)
	})

	t.Run(`Authorization・Cookie・Set-Cookie ヘッダをマスクし、Cookies は空にする`, func(t *testing.T) {
		event := &sentry.Event{
			Request: &sentry.Request{
				Cookies: "session=abc123",
				Headers: map[string]string{
					"Authorization": "Bearer secret-token",
					"Cookie":        "session=abc123",
					"Set-Cookie":    "session=abc123; HttpOnly",
					"User-Agent":    "go-test",
				},
			},
		}

		got := sentryx.MaskEvent(event, nil)

		assert.Equal(t, "", got.Request.Cookies)
		assert.Equal(t, "***", got.Request.Headers["Authorization"])
		assert.Equal(t, "***", got.Request.Headers["Cookie"])
		assert.Equal(t, "***", got.Request.Headers["Set-Cookie"])
		assert.Equal(t, "go-test", got.Request.Headers["User-Agent"], "秘匿対象以外のヘッダはそのまま残す")
	})

	t.Run(`ヘッダ名の大文字小文字が異なっていてもマスクする`, func(t *testing.T) {
		event := &sentry.Event{
			Request: &sentry.Request{
				Headers: map[string]string{
					"authorization": "Bearer secret-token",
					"COOKIE":        "session=abc123",
				},
			},
		}

		got := sentryx.MaskEvent(event, nil)

		assert.Equal(t, "***", got.Request.Headers["authorization"])
		assert.Equal(t, "***", got.Request.Headers["COOKIE"])
	})

	t.Run(`旧実装では拾えなかったヘッダも部分一致でマスクする`, func(t *testing.T) {
		// 旧 ScrubEvent は Authorization/Cookie/Set-Cookie の完全一致 3 つだけを見ていたため、
		// これらは伏せられずに送信されていた。masking.IsSensitive の部分一致化で拾えるようになる。
		event := &sentry.Event{
			Request: &sentry.Request{
				Headers: map[string]string{
					"X-Auth-Token": "secret-token",
					"X-Api-Secret": "secret-value",
					"User-Agent":   "go-test",
				},
			},
		}

		got := sentryx.MaskEvent(event, nil)

		assert.Equal(t, "***", got.Request.Headers["X-Auth-Token"], `"token" を含むので拾える`)
		assert.Equal(t, "***", got.Request.Headers["X-Api-Secret"], `"secret" を含むので拾える`)
		assert.Equal(t, "go-test", got.Request.Headers["User-Agent"])
	})

	t.Run(`従来は素通りしていた X-Api-Key もマスクする`, func(t *testing.T) {
		// masking.IsSensitive はキー名からハイフン・アンダースコアを除いてから判定するため、
		// ハイフン区切りの実運用ヘッダ "X-Api-Key" も "apikey" に一致してマスクされる。
		event := &sentry.Event{
			Request: &sentry.Request{
				Headers: map[string]string{
					"X-Api-Key": "secret-value",
				},
			},
		}

		got := sentryx.MaskEvent(event, nil)

		assert.Equal(t, "***", got.Request.Headers["X-Api-Key"])
	})

	t.Run(`QueryString の秘密情報にあたるキーだけを伏せる`, func(t *testing.T) {
		event := &sentry.Event{
			Request: &sentry.Request{
				QueryString: "q=aspirin&token=secret",
			},
		}

		got := sentryx.MaskEvent(event, nil)

		masked, err := url.ParseQuery(got.Request.QueryString)
		require.NoError(t, err)
		assert.Equal(t, url.Values{
			"q":     {"aspirin"},
			"token": {"***"},
		}, masked)
	})

	t.Run(`QueryString が空なら空のまま`, func(t *testing.T) {
		event := &sentry.Event{
			Request: &sentry.Request{QueryString: ""},
		}

		got := sentryx.MaskEvent(event, nil)

		assert.Equal(t, "", got.Request.QueryString)
	})

	t.Run(`解釈できない QueryString は漏らさず捨てる`, func(t *testing.T) {
		event := &sentry.Event{
			Request: &sentry.Request{
				// Go 1.17 以降、セミコロン区切りは不正なクエリとしてパースエラーになる
				QueryString: "a=1;b=2",
			},
		}

		got := sentryx.MaskEvent(event, nil)

		assert.Equal(t, "***", got.Request.QueryString, "解釈できないクエリは内容ごと伏せて捨てる")
	})
}
