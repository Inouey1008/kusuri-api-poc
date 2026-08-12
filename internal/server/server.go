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

// 依存を組み立て、ルートと middleware を登録した Server を返す
func New(db *sql.DB) *Server {
	return newServer(
		[]Registerer{
			drug.NewHandler(drug.NewService(drug.NewRepository(db))),
		},
		[]Middleware{
			logging.Middleware,
		},
	)
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

	// 先に定義した middleware が外側 (先に実行される)
	handler := http.Handler(mux)
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return &Server{handler: handler}
}

func (server *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	server.handler.ServeHTTP(writer, request)
}
