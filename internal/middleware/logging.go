package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/labstack/echo/v4"
)

// ログに載せるボディの上限。これを超えるものは記録しない
const maxLoggedBodyBytes = 4 << 10

func Logging(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		requestID := resolveRequestID(c)

		attrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.String("method", c.Request().Method),
			slog.String("path", c.Request().URL.Path),
			slog.Any("query", sanitizeQuery(c.QueryParams())),
			slog.Any("headers", pickHeaders(c.Request().Header)),
		}

		if body, ok := readBodyForLog(c); ok {
			attrs = append(attrs, slog.Any("body", body))
		}

		slog.LogAttrs(c.Request().Context(), slog.LevelInfo, "request", attrs...)

		err := next(c)

		if err != nil {
			// ステータスコードの確定は HTTPErrorHandler が行うが、これはミドルウェアの外側で走る
			// ここでエラーとして確定させないと、ログには 既定値の 200 のステータスコードが出力され、
			// エラーとログでステータスコードが異なってしまう
			c.Error(err)
		}

		status := c.Response().Status
		responseAttrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
		}

		if err != nil {
			responseAttrs = append(responseAttrs, slog.String("error", err.Error()))
		}

		slog.LogAttrs(c.Request().Context(), levelOf(status), "response", responseAttrs...)

		return err
	}
}

// ステータスコードからログレベルを決める
func levelOf(status int) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case status >= http.StatusBadRequest:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func resolveRequestID(c echo.Context) string {
	if lambdaContext, ok := core.GetRuntimeContextFromContextV2(c.Request().Context()); ok && lambdaContext != nil {
		return lambdaContext.AwsRequestID
	}

	if id := c.Request().Header.Get(echo.HeaderXRequestID); id != "" {
		return id
	}

	return randomID()
}

func randomID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "unknown"
	}

	return hex.EncodeToString(buf)
}

func readBodyForLog(c echo.Context) (any, bool) {
	request := c.Request()
	if request.Body == nil || request.Body == http.NoBody {
		return nil, false
	}

	if !strings.HasPrefix(request.Header.Get(echo.HeaderContentType), echo.MIMEApplicationJSON) {
		return nil, false
	}

	// 上限を 1 バイト超えて読み、超過を判定できるようにする
	raw, err := io.ReadAll(io.LimitReader(request.Body, maxLoggedBodyBytes+1))

	// 読んだ分を戻さないとハンドラがボディを受け取れない
	request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(raw), request.Body))

	if err != nil || len(raw) == 0 || len(raw) > maxLoggedBodyBytes {
		return nil, false
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, false
	}

	return sanitizeJSON(decoded), true
}
