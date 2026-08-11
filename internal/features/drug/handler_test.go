package drug_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/router"
)

// stubSvc は drug.NewHandler が要求するユースケース境界を満たすテスト用スタブ。
type stubSvc struct {
	searchResult []drug.Drug
	searchErr    error
	findResult   *drug.Drug
	findErr      error
}

func (s *stubSvc) Search(ctx context.Context, q string) ([]drug.Drug, error) {
	return s.searchResult, s.searchErr
}

func (s *stubSvc) GetByYJCode(ctx context.Context, yjCode string) (*drug.Drug, error) {
	return s.findResult, s.findErr
}

// newRouter はテスト用のスタブを使ったルーターを組み立てるヘルパ。
func newRouter(svc *stubSvc) http.Handler {
	return router.New(drug.NewHandler(svc))
}

func TestHandler_Search(t *testing.T) {
	stub := &stubSvc{
		searchResult: []drug.Drug{
			{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
			{YJCode: "2189018F1094", Name: "エゼチミブ錠10mg「YD」"},
		},
	}
	r := newRouter(stub)

	req := httptest.NewRequest(http.MethodGet, "/drugs?q=エゼチミブ", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body struct {
		Total int `json:"total"`
		Items []struct {
			YJCode string `json:"yjCode"`
			Name   string `json:"name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Total != 2 {
		t.Errorf("total = %d, want 2", body.Total)
	}
	if len(body.Items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(body.Items))
	}
	if body.Items[0].YJCode != "2189018F1043" {
		t.Errorf("items[0].yjCode = %q, want 2189018F1043", body.Items[0].YJCode)
	}
}

func TestHandler_GetByYJCode_Found(t *testing.T) {
	stub := &stubSvc{
		findResult: &drug.Drug{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
	}
	r := newRouter(stub)

	// router 経由でパスパラメータ {yjCode} が handler に届くことを確認
	req := httptest.NewRequest(http.MethodGet, "/drugs/2189018F1043", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var body struct {
		YJCode string `json:"yjCode"`
		Name   string `json:"name"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.YJCode != "2189018F1043" {
		t.Errorf("yjCode = %q, want 2189018F1043", body.YJCode)
	}
	if body.Name != "エゼチミブ錠10mg「JG」" {
		t.Errorf("name = %q, want エゼチミブ錠10mg「JG」", body.Name)
	}
}

func TestHandler_GetByYJCode_NotFound(t *testing.T) {
	stub := &stubSvc{findResult: nil, findErr: nil}
	r := newRouter(stub)

	req := httptest.NewRequest(http.MethodGet, "/drugs/0000000X0000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error != "対象が見つかりません" {
		t.Errorf("error = %q, want \"対象が見つかりません\"", body.Error)
	}
}

// validationResponse はバリデーションエラーレスポンスの形状。
type validationResponse struct {
	Error   string `json:"error"`
	Details []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"details"`
}

func TestHandler_Search_Validation(t *testing.T) {
	stub := &stubSvc{}
	r := newRouter(stub)

	// q が 101 文字 (ASCII) → 400
	longQ := ""
	for range 101 {
		longQ += "a"
	}

	req := httptest.NewRequest(http.MethodGet, "/drugs?q="+longQ, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}

	var body validationResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Error != "入力内容に誤りがあります" {
		t.Errorf("error = %q, want \"入力内容に誤りがあります\"", body.Error)
	}
	if len(body.Details) == 0 {
		t.Fatal("details is empty, want at least 1 entry")
	}
	if body.Details[0].Field != "q" {
		t.Errorf("details[0].field = %q, want \"q\"", body.Details[0].Field)
	}
}

func TestHandler_GetByYJCode_Validation(t *testing.T) {
	stub := &stubSvc{}
	r := newRouter(stub)

	cases := []struct {
		name    string
		yjCode  string
		wantErr string // details[0].field
	}{
		{
			name:    "短すぎるコード (3文字)",
			yjCode:  "abc",
			wantErr: "yjCode",
		},
		{
			name:    "12桁でも非英数記号含む",
			yjCode:  "12345678901!",
			wantErr: "yjCode",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/drugs/"+tc.yjCode, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", w.Code)
			}

			var body validationResponse
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if body.Error != "入力内容に誤りがあります" {
				t.Errorf("error = %q, want \"入力内容に誤りがあります\"", body.Error)
			}
			if len(body.Details) == 0 {
				t.Fatal("details is empty, want at least 1 entry")
			}
			if body.Details[0].Field != tc.wantErr {
				t.Errorf("details[0].field = %q, want %q", body.Details[0].Field, tc.wantErr)
			}
		})
	}
}
