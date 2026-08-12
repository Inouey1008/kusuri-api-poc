package testdb

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"

	"github.com/inouey1008/kusuri-api-poc/internal/sqlite"
)

// assets/gen.sql と同一のスキーマ・データ
const schema = `
CREATE TABLE drug (
  yj_code         TEXT PRIMARY KEY,
  name            TEXT NOT NULL,
  name_normalized TEXT NOT NULL
);
INSERT INTO drug VALUES
 ('2189018F1043','エゼチミブ錠10mg「JG」','エゼチミブ錠10mgjg'),
 ('2189018F1094','エゼチミブ錠10mg「YD」','エゼチミブ錠10mgyd'),
 ('1149037F2093','セレコキシブ錠100mg「サワイ」','セレコキシブ錠100mgサワイ');
CREATE INDEX idx_norm ON drug(name_normalized);
`

// 一時 SQLite ファイルを作成し、書き込み可能接続でスキーマとデータを投入して閉じる
func Setup(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")

	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)

	_, err = db.Exec(schema)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	return path
}

// 読み取り専用で接続した DB を返す
func Connect(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sqlite.Connect(context.Background(), Setup(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	return db
}
