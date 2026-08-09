// Package sqlite は SQLite への接続を開く共通処理を提供する。
// モジュール横断で使うのは接続の開き方だけで、クエリは各モジュールが持つ。
package sqlite

import (
	"context"
	"database/sql"

	// modernc.org/sqlite は純Go実装のため cgo 不要。Lambda 向けにも同じバイナリが使える。
	_ "modernc.org/sqlite"
)

// Connect は読み取り専用 (immutable=1) で開き、疎通確認まで行って *sql.DB を返す。
// immutable=1 はデプロイ先の読み取り専用 FS (/var/task) への対応に必要。
// PingContext でファイルオープンまで確認し、DB 欠損・破損を起動時に早期検出する (fail-fast)。
func Connect(ctx context.Context, dbPath string) (*sql.DB, error) {
	dsn := "file:" + dbPath + "?mode=ro&immutable=1"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// 起動時に実接続まで確立し、DB 欠損・破損を早期に検出する (fail-fast)。
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
