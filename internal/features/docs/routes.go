package docs

import "github.com/labstack/echo/v4"

func (handler *Handler) Register(e *echo.Echo) {
	e.GET("/openapi.yaml", handler.Spec)
	e.GET("/docs", handler.Page)
	e.GET("/docs/redoc.standalone.js", handler.Script)
}
