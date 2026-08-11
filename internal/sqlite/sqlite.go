package sqlite

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

// immutable=1 (Lambda の読み取り専用 FS 対応) で DB を開き、疎通確認まで行う
func Connect(ctx context.Context, dbPath string) (*sql.DB, error) {
	dsn := "file:" + dbPath + "?mode=ro&immutable=1"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// 起動時に実接続まで確立し、DB 欠損・破損を早期に検出する (fail-fast)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
