package drug_test

import (
	"context"
	"errors"
	"testing"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
)

// stubRepo は drug.Repository を満たすテスト用スタブ。
type stubRepo struct {
	searchResult []drug.Drug
	searchErr    error
	searchQuery  string // 受け取ったクエリを記録
}

func (s *stubRepo) Search(ctx context.Context, q string) ([]drug.Drug, error) {
	s.searchQuery = q
	return s.searchResult, s.searchErr
}

func TestService_Search(t *testing.T) {
	want := []drug.Drug{
		{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
		{YJCode: "2189018F1094", Name: "エゼチミブ錠10mg「YD」"},
	}
	stub := &stubRepo{searchResult: want}
	svc := drug.NewService(stub)

	got, err := svc.Search(context.Background(), "エゼチミブ")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// クエリがスタブに正しく届いているか
	if stub.searchQuery != "エゼチミブ" {
		t.Errorf("searchQuery = %q, want %q", stub.searchQuery, "エゼチミブ")
	}
	// 結果がそのまま返るか
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i, d := range got {
		if d.YJCode != want[i].YJCode {
			t.Errorf("got[%d].YJCode = %q, want %q", i, d.YJCode, want[i].YJCode)
		}
	}
}

func TestService_Search_Error(t *testing.T) {
	stub := &stubRepo{searchErr: errors.New("db error")}
	svc := drug.NewService(stub)

	_, err := svc.Search(context.Background(), "エゼチミブ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
