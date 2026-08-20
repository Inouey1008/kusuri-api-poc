package server

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
)

// ハンドラが返した error を JSON へ変換する。
// 応答の組み立てをここに集約し、ハンドラは error を返すだけにする。
// 判別できない error は 500 に倒す。内容は Logging ミドルウェアが記録する。
func errorHandler(err error, c echo.Context) {
	// Logging ミドルウェアが既に書いている場合がある
	if c.Response().Committed {
		return
	}

	target := errorx.Internal

	var appError errorx.Errorx
	var echoError *echo.HTTPError

	switch {
	case errors.As(err, &appError):
		target = appError

	case errors.As(err, &echoError):
		// Echo が返す 404 や 405 をアプリのエラー形式へ揃える
		target = errorx.Errorx{
			Status:  echoError.Code,
			Code:    statusCodeName(echoError.Code),
			Message: http.StatusText(echoError.Code),
		}
	}

	if writeErr := c.JSON(target.Status, target); writeErr != nil {
		slog.ErrorContext(c.Request().Context(), "failed to write error response", slog.Any("error", writeErr))
	}
}

// 404 → "NOT_FOUND" のように、status から言語非依存の識別子を作る
func statusCodeName(status int) string {
	text := http.StatusText(status)
	if text == "" {
		return "HTTP_ERROR"
	}

	return strings.ToUpper(strings.ReplaceAll(text, " ", "_"))
}
