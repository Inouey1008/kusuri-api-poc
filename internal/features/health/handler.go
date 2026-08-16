package health

import (
	"net/http"

	"github.com/inouey1008/kusuri-api-poc/internal/httpx"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// DB は起動時に sqlite.Connect が疎通確認済みのため、ここでは接続を見ない
func (handler *Handler) Check(writer http.ResponseWriter, request *http.Request) {
	httpx.WriteJSON(writer, http.StatusOK, healthResponse{Status: "ok"})
}
