package drug

import "net/http"

// Register は自身のルートを mux に登録する。router 側はモジュールを知らなくてよい。
func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /drugs", handler.Search)
}
