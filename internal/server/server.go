package server

import (
	"database/sql"

	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
	"github.com/inouey1008/kusuri-api-poc/internal/features/docs"
	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/features/health"
	"github.com/inouey1008/kusuri-api-poc/internal/middleware"
	"github.com/inouey1008/kusuri-api-poc/internal/sentryx"
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

	middlewares := []echo.MiddlewareFunc{
		sentryecho.New(sentryecho.Options{
			Repanic:         true,
			WaitForDelivery: true, // Lambda はコンテナ凍結前に配送を待つ必要があるため明示する
			Timeout:         sentryx.FlushTimeout,
		}),
		middleware.Logging,
		middleware.Recovery, // !! Recovery を内側に置くことで、panic した要求も 500 として Logging に記録される
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
