package drug_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/server"
)

type stubService struct {
	searchResult []drug.Drug
	searchErr    error
}

func (s *stubService) Search(ctx context.Context, q string) ([]drug.Drug, error) {
	return s.searchResult, s.searchErr
}

func callSearch(t *testing.T, service *stubService, q string) *httptest.ResponseRecorder {
	t.Helper()

	e := server.NewWith([]server.Registerer{drug.NewHandler(service)}, nil)

	target := "/drugs?q=" + url.QueryEscape(q)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))

	return recorder
}

func TestHandler_Search(t *testing.T) {
	t.Run(`一致した薬を JSON で返す`, func(t *testing.T) {
		stub := &stubService{
			searchResult: []drug.Drug{
				{YJCode: "2189018F1043", Name: "エゼチミブ錠10mg「JG」"},
				{YJCode: "2189018F1094", Name: "エゼチミブ錠10mg「YD」"},
			},
		}

		recorder := callSearch(t, stub, "エゼチミブ")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("Content-Type"), "application/json")

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
	})

	t.Run(`q が長すぎる場合は 400 と対象フィールドを返す`, func(t *testing.T) {
		recorder := callSearch(t, &stubService{}, strings.Repeat("a", 101))

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
	})
}
