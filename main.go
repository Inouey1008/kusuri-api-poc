package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
	"github.com/inouey1008/kusuri-api-poc/internal/logx"
	"github.com/inouey1008/kusuri-api-poc/internal/server"
	"github.com/inouey1008/kusuri-api-poc/internal/sqlite"
)

const (
	dbPath          = "./assets/master.db"
	shutdownTimeout = 10 * time.Second
)

func main() {
	// Logger 設定
	logx.Init()

	// 環境変数の検証
	cfg, err := config.Load()
	if err != nil {
		slog.Error("invalid configuration", slog.Any("error", err))
		os.Exit(1)
	}

	// DB との接続検証
	db, err := sqlite.Connect(context.Background(), dbPath)
	if err != nil {
		slog.Error("failed to connect database", slog.Any("error", err), slog.String("path", dbPath))
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	e := server.New(db)

	if cfg.OnLambda() {
		// Function URLs は API Gateway v2 ペイロード形式のため NewV2 を使う。
		lambda.Start(echoadapter.NewV2(e).ProxyWithContext)
		return
	}

	if err := listenAndServe(e, cfg.Port); err != nil {
		slog.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}
}

// SIGINT / SIGTERM を受けたら処理中のリクエストを待ってから終了する
func listenAndServe(e *echo.Echo, port string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		slog.Info("listening", slog.String("port", port))
		if err := e.Start(":" + port); !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return e.Shutdown(shutdownCtx)
}
