// Package sentryx は Sentry 連携の共通処理をまとめる。
// 初期化 (Init) と、capture の直後に Flush を呼ぶ処理 (Report) を集約する。
// 後者をまとめているのは、呼び出し側で Flush を書き忘れると
// Lambda のコンテナ凍結でイベントが黙って消えてしまうため。
//
// 依存は echo / sentryecho と internal/config・internal/errorx・internal/masking に限る。
// internal/middleware・internal/server からは参照される側なので、
// 循環参照を避けるためそれらには依存しない。
package sentryx

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
	"github.com/inouey1008/kusuri-api-poc/internal/masking"
)

const (
	// FlushTimeout は送信完了を待つ上限時間。
	FlushTimeout = 2 * time.Second

	// maxTagValueLength は Sentry のタグの値の上限文字数。
	maxTagValueLength = 200
)

// Sentry を初期化する
// DSN が空なら、未設定とみなし、連携せず続行する
// DSN が不正なら error を返す
func Init(cfg config.Config) error {
	if cfg.SentryDSN == "" {
		msg := "sentry is disabled"
		args := []any{slog.String("reason", "SENTRY_DSN is not set")}
		if cfg.IsLocal() {
			// local では Info に留める
			slog.Info(msg, args...)
		} else {
			// デプロイ先では監視が効いていない状態に気づけるよう Warn で残す
			slog.Warn(msg, args...)
		}
		return nil
	}

	// Release は指定しない。SDK がバイナリに埋め込まれたコミット ID (vcs.revision) を自動で使う。
	return sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.Environment,
		TracesSampleRate: cfg.SentryTracesSampleRate,
		BeforeSend:       maskEvent,
	})
}

// Report は error を Sentry へ送信し、送信完了まで待つ。送信はここに一本化している。
func Report(c echo.Context, err error) {
	// echo が自動で返した 4xx (ルートなしの 404、メソッド違いの 405、BasicAuth の 401 など) は送らない。
	// bot や利用者の誤操作由来で問い合わせの対象にならず、クォータを無駄に消費するため。
	// アプリが意図して返す 4xx は errorx で返すこと (そちらは送る)。
	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) && echoErr.Code >= http.StatusBadRequest && echoErr.Code < http.StatusInternalServerError {
		return
	}

	hub := sentryecho.GetHubFromContext(c)
	if hub == nil { // DSN 未設定など Sentry が無効なとき
		return
	}

	// WithScope 内で設定したタグとレベルは、抜けると元に戻る (次の送信に漏れない)
	hub.WithScope(func(scope *sentry.Scope) {
		// ログと同じ request_id を載せ、Sentry で ID を検索 → CloudWatch を同じ ID で検索、と辿れるようにする
		if id := c.Response().Header().Get(echo.HeaderXRequestID); id != "" {
			scope.SetTag("request_id", truncate(id, maxTagValueLength))
		}

		// panic は致命的な不具合なので、通常の error (Error レベル) と区別できるよう Fatal で送る
		var panicErr *errorx.Panic
		if errors.As(err, &panicErr) {
			scope.SetLevel(sentry.LevelFatal)
		}

		hub.CaptureException(err)
	})

	// CaptureException は送信を予約するだけで、実際の送信は裏の goroutine が行う。
	// Lambda はレスポンスを返すとコンテナを凍結するので、送り終える前に止まらないようここで待つ。
	hub.Flush(FlushTimeout)
}

// truncate は s を最大 limit 文字 (rune 単位) に切り詰める。
// request_id は外部から来る値で長さの保証がないため。rune 単位なのはマルチバイト文字の途中で切らないため。
func truncate(s string, limit int) string {
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}

	return string(runes[:limit])
}

// maskEvent は Authorization や API キーなど機微な値を送信前に伏せる (BeforeSend)。
// 判定は masking に寄せ、ログ側と基準がズレないようにしている。
func maskEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	// リクエスト情報がないイベントの場合は、何もしない
	if event.Request == nil {
		return event
	}

	// Cookie
	// Sentry 側で自動的に送信されるので、ここでは空にする
	event.Request.Cookies = ""

	// Header
	// Authorization や API キーなどのマスキング
	for name := range event.Request.Headers {
		if masking.IsSensitive(name) {
			event.Request.Headers[name] = masking.MaskedValue
		}
	}

	// QueryString
	// Authorization や API キーなどのマスキング
	event.Request.QueryString = maskQueryString(event.Request.QueryString)

	return event
}

// マスクした結果でクエリ文字列を組み立て直す。
// 解釈できないクエリは、漏らすより捨てる方を選ぶ。
func maskQueryString(raw string) string {
	if raw == "" {
		return ""
	}

	values, err := url.ParseQuery(raw)
	// ParseQuery が失敗するのは、クエリ文字列が不正な場合のみ。
	// その場合は、捨てる方が安全なので、マスク済みの値を返す。
	if err != nil {
		return masking.MaskedValue
	}

	return url.Values(masking.Values(values)).Encode()
}
