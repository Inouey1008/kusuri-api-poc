package health

import "github.com/labstack/echo/v4"

func (handler *Handler) Register(e *echo.Echo) {
	e.GET("/health", handler.Check)
}
