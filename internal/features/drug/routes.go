package drug

import "net/http"

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /drugs", handler.Search)
}
