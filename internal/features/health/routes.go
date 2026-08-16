package health

import "net/http"

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", handler.Check)
}
