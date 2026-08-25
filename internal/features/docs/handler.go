package docs

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/inouey1008/kusuri-api-poc/internal/config"
)

//go:embed openapi.yaml
var spec []byte

//go:embed index.html
var page []byte

//go:embed redoc.standalone.js
var script []byte

type Handler struct {
	config config.Config
}

func NewHandler(config config.Config) *Handler {
	return &Handler{config: config}
}

func (handler *Handler) authorize(user, password string, _ echo.Context) (bool, error) {
	return user == handler.config.DocsUser && password == handler.config.DocsPassword, nil
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
