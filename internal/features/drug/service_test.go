package drug_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
)

type stubRepository struct {
	searchResult []drug.Drug
	searchErr    error
	searchQuery  string
}

func (s *stubRepository) FindByName(ctx context.Context, q string) ([]drug.Drug, error) {
	s.searchQuery = q
	return s.searchResult, s.searchErr
}

func TestService_Search(t *testing.T) {
	t.Run(`リポジトリの結果をそのまま返す`, func(t *testing.T) {
		expected := []drug.Drug{
			{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
			{YJCode: "2189018F1094", Name: "エゼチミブ錠10mg「YD」"},
		}
		service := drug.NewService(&stubRepository{searchResult: expected})

		got, err := service.Search(context.Background(), "エゼチミブ")

		require.NoError(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run(`検索語をリポジトリへ渡す`, func(t *testing.T) {
		stub := &stubRepository{}
		service := drug.NewService(stub)

		_, err := service.Search(context.Background(), "エゼチミブ")

		require.NoError(t, err)
		assert.Equal(t, "エゼチミブ", stub.searchQuery)
	})

	t.Run(`リポジトリが失敗したらエラーを返す`, func(t *testing.T) {
		service := drug.NewService(&stubRepository{searchErr: errors.New("db error")})

		_, err := service.Search(context.Background(), "エゼチミブ")

		assert.Error(t, err)
	})
}
