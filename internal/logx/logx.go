package logx

import (
	"log/slog"
	"os"
)

// ログを JSON で構造化して出力するよう設定する
func Init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
