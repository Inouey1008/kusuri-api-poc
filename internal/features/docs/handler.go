package docs

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed openapi.yaml
var spec []byte

//go:embed index.html
var page []byte

//go:embed redoc.standalone.js
var script []byte

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (handler *Handler) Spec(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/yaml", spec)
}

func (handler *Handler) Page(c echo.Context) error {
	return c.HTMLBlob(http.StatusOK, page)
}

func (handler *Handler) Script(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/javascript", script)
}
