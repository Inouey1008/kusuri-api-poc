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

func (s *stubRepository) Search(ctx context.Context, q string) ([]drug.Drug, error) {
	s.searchQuery = q
	return s.searchResult, s.searchErr
}

func TestService_Search_Success(t *testing.T) {
	expected := []drug.Drug{
		{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
		{YJCode: "2189018F1094", Name: "エゼチミブ錠10mg「YD」"},
	}
	stub := &stubRepository{searchResult: expected}
	service := drug.NewService(stub)

	got, err := service.Search(context.Background(), "エゼチミブ")

	require.NoError(t, err)
	assert.Equal(t, "エゼチミブ", stub.searchQuery)
	assert.Equal(t, expected, got)
}

func TestService_Search_Error(t *testing.T) {
	stub := &stubRepository{searchErr: errors.New("db error")}
	service := drug.NewService(stub)

	_, err := service.Search(context.Background(), "エゼチミブ")

	assert.Error(t, err)
}
