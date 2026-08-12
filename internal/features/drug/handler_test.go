package drug_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/router"
)

type stubService struct {
	searchResult []drug.Drug
	searchErr    error
}

func (s *stubService) Search(ctx context.Context, q string) ([]drug.Drug, error) {
	return s.searchResult, s.searchErr
}

func newRouter(service *stubService) http.Handler {
	return router.New(drug.NewHandler(service))
}

func TestHandler_Search_Success(t *testing.T) {
	stub := &stubService{
		searchResult: []drug.Drug{
			{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
			{YJCode: "2189018F1094", Name: "エゼチミブ錠10mg「YD」"},
		},
	}
	handler := newRouter(stub)

	request := httptest.NewRequest(http.MethodGet, "/drugs?q=エゼチミブ", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var body struct {
		Total int `json:"total"`
		Items []struct {
			YJCode string `json:"yjCode"`
			Name   string `json:"name"`
		} `json:"items"`
	}
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))

	assert.Equal(t, 2, body.Total)
	require.Len(t, body.Items, 2)
	assert.Equal(t, "2189018F1043", body.Items[0].YJCode)
}

func TestHandler_Search_Validation(t *testing.T) {
	stub := &stubService{}
	handler := newRouter(stub)

	request := httptest.NewRequest(http.MethodGet, "/drugs?q="+strings.Repeat("a", 101), nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)

	var body struct {
		Error   string `json:"error"`
		Details []struct {
			Field   string `json:"field"`
			Message string `json:"message"`
		} `json:"details"`
	}
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))

	assert.Equal(t, "入力内容に誤りがあります", body.Error)
	require.NotEmpty(t, body.Details)
	assert.Equal(t, "q", body.Details[0].Field)
}
