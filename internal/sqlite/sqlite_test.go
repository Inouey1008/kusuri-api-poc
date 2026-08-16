package sqlite_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"

	"github.com/inouey1008/kusuri-api-poc/internal/sqlite"
)

// 一時 SQLite ファイルを作成し、テーブルを 1 つ作って閉じる
func setupDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")

	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)

	_, err = db.Exec("CREATE TABLE drug (yj_code TEXT PRIMARY KEY)")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	return path
}

func TestConnect_Success(t *testing.T) {
	db, err := sqlite.Connect(context.Background(), setupDB(t))

	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	assert.NoError(t, db.PingContext(context.Background()))
}

func TestConnect_Error(t *testing.T) {
	testCases := []struct {
		name string
		path string
	}{
		{
			name: `存在しないファイル`,
			path: filepath.Join(t.TempDir(), "missing.db"),
		},
		{
			name: `ディレクトリを指定`,
			path: t.TempDir(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			db, err := sqlite.Connect(context.Background(), testCase.path)

			assert.Error(t, err, "PingContext による fail-fast で失敗するべき")
			assert.Nil(t, db)
		})
	}
}

func TestConnect_ReadOnly(t *testing.T) {
	db, err := sqlite.Connect(context.Background(), setupDB(t))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.Exec("INSERT INTO drug VALUES ('9999999X9999')")

	assert.Error(t, err, "immutable=1 のため書き込みは失敗する")
}
