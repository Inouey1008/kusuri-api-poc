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
}

func (s *stubSvc) Search(ctx context.Context, q string) ([]drug.Drug, error) {
	return s.searchResult, s.searchErr
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

// validationResponse はバリデーションエラーレスポンスの形状。
type validationResponse struct {
	Error   string `json:"error"`
	Details []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	} `json:"details"`
}

// バリデーションが handler に配線され 400 を返すことだけを確認する。
// ルール自体の網羅は dto_test.go の TestSearchRequest_Validate が担う。
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
