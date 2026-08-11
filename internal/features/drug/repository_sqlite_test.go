package drug_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/sqlite"
)

// gen.sql と同一のスキーマ・データ。
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

// setupDB はテスト用の一時 SQLite ファイルを作成し、書き込み可能接続でスキーマとデータを投入して閉じる。
// 戻り値は db ファイルのパス。Open (immutable/ro) での検証はテスト側で行う。
func setupDB(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")

	rw, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatalf("setup open: %v", err)
	}
	if _, err := rw.Exec(testSchema); err != nil {
		rw.Close()
		t.Fatalf("setup exec: %v", err)
	}
	if err := rw.Close(); err != nil {
		t.Fatalf("setup close: %v", err)
	}
	return path
}

func TestSQLiteRepository_Search(t *testing.T) {
	path := setupDB(t)
	db, err := sqlite.Connect(context.Background(), path)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer db.Close()

	repo := drug.NewRepository(db)
	ctx := context.Background()

	cases := []struct {
		query    string
		wantLen  int
		wantCode string // 先頭件の yj_code。空文字はチェックしない。
	}{
		{"エゼチミブ", 2, "2189018F1043"},
		{"セレコキシブ", 1, "1149037F2093"},
		{"jg", 1, "2189018F1043"},
		{"", 3, ""},
		{"存在しない薬", 0, ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.query, func(t *testing.T) {
			got, err := repo.Search(ctx, tc.query)
			if err != nil {
				t.Fatalf("Search(%q): %v", tc.query, err)
			}
			if len(got) != tc.wantLen {
				t.Errorf("Search(%q): got %d items, want %d", tc.query, len(got), tc.wantLen)
			}
			// 0件のときも nil でなく空スライスであること
			if got == nil {
				t.Errorf("Search(%q): got nil, want []drug.Drug{}", tc.query)
			}
			if tc.wantCode != "" && len(got) > 0 && got[0].YJCode != tc.wantCode {
				t.Errorf("Search(%q): got[0].YJCode = %q, want %q", tc.query, got[0].YJCode, tc.wantCode)
			}
		})
	}
}

func TestSQLiteRepository_ReadOnly(t *testing.T) {
	path := setupDB(t)
	db, err := sqlite.Connect(context.Background(), path)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer db.Close()

	// immutable=1 が有効な場合、書き込みはエラーになる。
	_, err = db.Exec("INSERT INTO drug VALUES ('9999999X9999','テスト薬','テスト薬')")
	if err == nil {
		t.Fatal("expected error on INSERT to immutable=1 db, got nil")
	}
}
