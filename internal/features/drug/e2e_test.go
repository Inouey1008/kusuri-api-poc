package drug_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/router"
	"github.com/inouey1008/kusuri-api-poc/internal/sqlite"
)

// newHandler は実 temp DB を使ったフルスタック配線を組み上げ、http.Handler を返す。
// db は t.Cleanup で Close される。
func newHandler(t *testing.T) http.Handler {
	t.Helper()
	path := setupDB(t)

	db, err := sqlite.Connect(context.Background(), path)
	if err != nil {
		t.Fatalf("sqlite.Connect: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := drug.NewRepository(db)
	svc := drug.NewService(repo)
	h := drug.NewHandler(svc)
	return router.New(h)
}

// searchResponse は GET /drugs のレスポンス形状。
type searchResponse struct {
	Total int `json:"total"`
	Items []struct {
		YJCode string `json:"yjCode"`
		Name   string `json:"name"`
	} `json:"items"`
}

func TestE2E_Search(t *testing.T) {
	r := newHandler(t)

	cases := []struct {
		name      string
		q         string
		wantTotal int
		wantCodes []string // items に含まれるべき YJCode の集合 (順序不問)
	}{
		{
			name:      "エゼチミブで2件",
			q:         "エゼチミブ",
			wantTotal: 2,
			wantCodes: []string{"2189018F1043", "2189018F1094"},
		},
		{
			name:      "セレコキシブで1件",
			q:         "セレコキシブ",
			wantTotal: 1,
			wantCodes: []string{"1149037F2093"},
		},
		{
			name:      "空クエリで全3件",
			q:         "",
			wantTotal: 3,
			wantCodes: nil, // コード一致チェックは省略
		},
		{
			name:      "存在しない薬で0件",
			q:         "存在しない薬",
			wantTotal: 0,
			wantCodes: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/drugs?q="+tc.q, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			ct := rec.Header().Get("Content-Type")
			if ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			var resp searchResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}

			if resp.Total != tc.wantTotal {
				t.Errorf("total = %d, want %d", resp.Total, tc.wantTotal)
			}
			if len(resp.Items) != tc.wantTotal {
				t.Errorf("len(items) = %d, want %d", len(resp.Items), tc.wantTotal)
			}

			// 0件のとき items は nil でなく空配列であることも検証する
			if tc.wantTotal == 0 && resp.Items == nil {
				t.Error("items is nil, want empty array")
			}

			// 期待 YJCode が items に含まれるか確認する
			if len(tc.wantCodes) > 0 {
				codeSet := make(map[string]bool, len(resp.Items))
				for _, item := range resp.Items {
					codeSet[item.YJCode] = true
				}
				for _, code := range tc.wantCodes {
					if !codeSet[code] {
						t.Errorf("YJCode %q not found in items", code)
					}
				}
			}
		})
	}
}
