package server

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
	"github.com/inouey1008/kusuri-api-poc/internal/features/docs"
	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/features/health"
	"github.com/inouey1008/kusuri-api-poc/internal/middleware"
)

type Registerer interface {
	Register(e *echo.Echo)
}

func New(config config.Config, db *sql.DB) *echo.Echo {
	healthHandler := health.NewHandler()

	docsHandler := docs.NewHandler(config)

	drugRepository := drug.NewRepository(db)
	drugService := drug.NewService(drugRepository)
	drugHandler := drug.NewHandler(drugService)

	registerers := []Registerer{
		healthHandler,
		docsHandler,
		drugHandler,
	}

	// スライスの先頭ほど外側のミドルウェアになる
	middlewares := []echo.MiddlewareFunc{
		middleware.RequestID, // 最外側: 内側のすべてがレスポンスヘッダから request_id を読む
		middleware.Sentry,    // 内側の Recovery と errorHandler が hub を取り出す
		middleware.Logging,
		middleware.Recovery, // 最内側: panic を 500 に変え、外側の Logging に記録させる
	}

	return newServer(registerers, middlewares)
}

// テストからスタブを差し込むための入口
func NewWith(registerers []Registerer, middlewares []echo.MiddlewareFunc) *echo.Echo {
	return newServer(registerers, middlewares)
}

func newServer(registerers []Registerer, middlewares []echo.MiddlewareFunc) *echo.Echo {
	e := echo.New()

	// Lambda では標準出力がそのままログになるため、Echo の装飾を止める
	e.HideBanner = true
	e.HidePort = true

	e.HTTPErrorHandler = errorHandler
	e.Use(middlewares...)

	for _, registerer := range registerers {
		registerer.Register(e)
	}

	return e
}
