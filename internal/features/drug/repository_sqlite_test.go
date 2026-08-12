package drug_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/sqlite"
)

// gen.sql と同一のスキーマ・データ
const testSchema = `
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
func setupDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")

	db, err := sql.Open("sqlite", "file:"+path)
	require.NoError(t, err)

	_, err = db.Exec(testSchema)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	return path
}

// 読み取り専用で接続した DB を返す
func connectDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sqlite.Connect(context.Background(), setupDB(t))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	return db
}

func TestSQLiteRepository_FindByName(t *testing.T) {
	repository := drug.NewRepository(connectDB(t))

	testCases := []struct {
		name         string
		query        string
		expectedLen  int
		expectedCode string // 先頭件の yj_code。空文字なら検証しない
	}{
		{
			name:         `部分一致で複数件`,
			query:        "エゼチミブ",
			expectedLen:  2,
			expectedCode: "2189018F1043",
		},
		{
			name:         `部分一致で 1 件`,
			query:        "セレコキシブ",
			expectedLen:  1,
			expectedCode: "1149037F2093",
		},
		{
			name:         `正規化された名前で一致 (小文字)`,
			query:        "jg",
			expectedLen:  1,
			expectedCode: "2189018F1043",
		},
		{
			name:        `空クエリは全件`,
			query:       "",
			expectedLen: 3,
		},
		{
			name:        `一致なしは 0 件`,
			query:       "存在しない薬",
			expectedLen: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := repository.FindByName(context.Background(), testCase.query)

			require.NoError(t, err)
			assert.NotNil(t, got) // 0 件でも nil ではなく空配列
			require.Len(t, got, testCase.expectedLen)

			if testCase.expectedCode != "" {
				assert.Equal(t, testCase.expectedCode, got[0].YJCode)
			}
		})
	}
}
