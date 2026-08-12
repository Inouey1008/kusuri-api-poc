package drug_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/server"
)

// 実 temp DB を使い、本番と同じ配線でフルスタックを組み上げる
func newHandler(t *testing.T) http.Handler {
	t.Helper()
	return server.New(connectDB(t))
}

func request(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	newHandler(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

	return recorder
}

// DB から取得した値が DTO を経て JSON になるまでを検証する
// 検索条件そのものの網羅は repository_sqlite_test.go が担う
func TestE2E_Search_Success(t *testing.T) {
	recorder := request(t, "/drugs?q=エゼチミブ")

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

	require.Len(t, body.Items, 2)
	assert.Equal(t, 2, body.Total)
	assert.Equal(t, "2189018F1043", body.Items[0].YJCode)
	assert.Equal(t, "エゼチミブ錠10mg「JG」", body.Items[0].Name)
}

// 0 件でも items が null にならないこと (JSON の形が崩れない)
func TestE2E_Search_Empty(t *testing.T) {
	recorder := request(t, "/drugs?q=存在しない薬")

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"total":0,"items":[]}`, recorder.Body.String())
}

// バリデーションで弾かれた場合に errorx の応答が返ること
func TestE2E_Search_Validation(t *testing.T) {
	recorder := request(t, "/drugs?q="+strings.Repeat("a", 101))

	require.Equal(t, http.StatusBadRequest, recorder.Code)

	var body struct {
		Code    string `json:"code"`
		Message string `json:"error"`
		Details []struct {
			Field string `json:"field"`
		} `json:"details"`
	}
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&body))

	assert.Equal(t, "VALIDATION_FAILED", body.Code)
	assert.NotEmpty(t, body.Message)
	require.NotEmpty(t, body.Details)
	assert.Equal(t, "q", body.Details[0].Field)
}

// 未登録のパスは server の mux が 404 を返す
func TestE2E_NotFound(t *testing.T) {
	recorder := request(t, "/unknown")

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
