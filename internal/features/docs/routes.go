package docs

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (handler *Handler) Register(e *echo.Echo) {
	var auth []echo.MiddlewareFunc
	if !handler.config.IsLocal() {
		auth = append(auth, middleware.BasicAuth(handler.authorize))
	}

	e.GET("/openapi.yaml", handler.Spec, auth...)
	e.GET("/docs", handler.Page, auth...)
	e.GET("/docs/redoc.standalone.js", handler.Script, auth...)
}
