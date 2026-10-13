package middleware

import (
	sentryecho "github.com/getsentry/sentry-go/echo"

	"github.com/inouey1008/kusuri-api-poc/internal/sentryx"
)

// Sentry はリクエストごとに複製した *sentry.Hub を echo.Context に載せる。
// 内側の Recovery と errorHandler (sentryx.Report) は、ここで載せた hub を取り出して送信する。
var Sentry = sentryecho.New(sentryecho.Options{
	Repanic:         true,
	WaitForDelivery: true, // Lambda はコンテナ凍結前に配送を待つ必要があるため明示する
	Timeout:         sentryx.FlushTimeout,
})
