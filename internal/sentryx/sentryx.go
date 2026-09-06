// Package sentryx は Sentry 連携の共通処理をまとめる。
// middleware / server から参照される一方でそれらに依存しないよう、独立したパッケージにしている。
package sentryx

import (
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

const (
	// FlushTimeout はイベント送信を待つ上限時間。Lambda はレスポンス後にコンテナが凍結され得るため、
	// ハンドラを抜ける前に同期的に flush してこの時間だけ待つ。
	FlushTimeout = 2 * time.Second

	// ReportedKey は Recovery が panic を Sentry へ送信済みであることを示す echo.Context のキー。
	// error_handler での二重送信を防ぐために使う。
	ReportedKey = "sentry_reported"
)

// ScrubEvent は Authorization / Cookie など機微なヘッダを送信前にマスクする (BeforeSend)。
func ScrubEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	if event.Request == nil {
		return event
	}

	event.Request.Cookies = ""
	for name := range event.Request.Headers {
		switch strings.ToLower(name) {
		case "authorization", "cookie", "set-cookie":
			event.Request.Headers[name] = "***"
		}
	}

	return event
}
