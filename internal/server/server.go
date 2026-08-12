package server

import (
	"database/sql"
	"net/http"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/logging"
)

type Middleware func(http.Handler) http.Handler

// 自身のルートを mux に登録できるモジュール
type Registerer interface {
	Register(mux *http.ServeMux)
}

type Server struct {
	handler http.Handler
}

func New(db *sql.DB) *Server {
	drugRepository := drug.NewRepository(db)
	drugService := drug.NewService(drugRepository)
	drugHandler := drug.NewHandler(drugService)

	registerers := []Registerer{
		drugHandler,
	}
	middlewares := []Middleware{
		logging.Middleware,
	}

	return newServer(registerers, middlewares)
}

// テストからスタブを差し込むための入口
func NewWith(registerers []Registerer, middlewares []Middleware) *Server {
	return newServer(registerers, middlewares)
}

func newServer(registerers []Registerer, middlewares []Middleware) *Server {
	mux := http.NewServeMux()
	for _, registerer := range registerers {
		registerer.Register(mux)
	}

	// middlewares[0] から順に実行されるよう、末尾から包んでいく
	// 例: {A, B} の場合 A(B(mux)) となり、A → B → mux の順に処理される
	handler := http.Handler(mux)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return &Server{handler: handler}
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.handler.ServeHTTP(writer, request)
}
