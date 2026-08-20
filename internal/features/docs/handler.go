package docs

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed index.html
var page []byte

//go:embed redoc.standalone.js
var script []byte

type Handler struct {
	spec []byte
}

func NewHandler(spec []byte) *Handler {
	return &Handler{spec: spec}
}

func (handler *Handler) Spec(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/yaml", handler.spec)
}

func (handler *Handler) Page(c echo.Context) error {
	return c.HTMLBlob(http.StatusOK, page)
}

func (handler *Handler) Script(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/javascript", script)
}
