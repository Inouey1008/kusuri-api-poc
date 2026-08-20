package integration_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/test/testrequest"
)

func TestSearch(t *testing.T) {
	t.Run(`一致した薬を返す`, func(t *testing.T) {
		recorder := testrequest.Get(t, "/drugs?q=エゼチミブ")

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
	})

	t.Run(`0 件でも items は空配列になる`, func(t *testing.T) {
		recorder := testrequest.Get(t, "/drugs?q=存在しない薬")

		require.Equal(t, http.StatusOK, recorder.Code)
		assert.JSONEq(t, `{"total":0,"items":[]}`, recorder.Body.String(), "null になると JSON の形が崩れる")
	})

	t.Run(`q が長すぎる場合は 400 と対象フィールドを返す`, func(t *testing.T) {
		recorder := testrequest.Get(t, "/drugs?q="+strings.Repeat("a", 101))

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
	})
}
