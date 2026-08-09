package router

import (
	"log"
	"net/http"
	"time"
)

// Registerer は自身のルートを mux に登録できるモジュールを表す。
// router はモジュールを import しないため、追加時にこのファイルを変更する必要がない。
type Registerer interface {
	Register(mux *http.ServeMux)
}

// New は各モジュールのルートを登録し、ログ middleware でラップして返す。
func New(rs ...Registerer) http.Handler {
	mux := http.NewServeMux()
	for _, r := range rs {
		r.Register(mux)
	}
	return logMiddleware(mux)
}

// logMiddleware はメソッド・パス・所要時間をログ出力する。
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
