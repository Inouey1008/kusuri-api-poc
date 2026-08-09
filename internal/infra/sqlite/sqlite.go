// Package sqlite は SQLite への接続を開く共通処理を提供する。
// モジュール横断で使うのは接続の開き方だけで、クエリは各モジュールが持つ。
package sqlite

import (
	"database/sql"

	// modernc.org/sqlite は純Go実装のため cgo 不要。Lambda 向けにも同じバイナリが使える。
	_ "modernc.org/sqlite"
)

// Open はデプロイ先が読み取り専用のため immutable=1 を指定して SQLite を開く。
// immutable=1 により WAL/ジャーナル操作をスキップし、コールドスタート時間を短縮できる。
func Open(dbPath string) (*sql.DB, error) {
	dsn := "file:" + dbPath + "?mode=ro&immutable=1"
	return sql.Open("sqlite", dsn)
}
