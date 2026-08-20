package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/labstack/echo/v4"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
)

func Recovery(next echo.HandlerFunc) echo.HandlerFunc {
	// 戻り値に名前を付け、defer から差し替えられるようにする
	return func(c echo.Context) (err error) {
		// panic は通常の戻り値の経路を通らないので、defer の中でしか捕まえられない
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			slog.ErrorContext(c.Request().Context(), "panic",
				slog.Any("value", recovered),
				slog.String("stack", string(debug.Stack())),
			)

			// panic を捕まえて 500 に変換する
			// 捕まえないと panic はハンドラの外まで抜け、Lambda が 502 を返す
			err = errorx.Internal
		}()

		return next(c)
	}
}
