package health

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// DB は起動時に sqlite.Connect が疎通確認済みのため、ここでは接続を見ない
func (handler *Handler) Check(c echo.Context) error {
	return c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}
