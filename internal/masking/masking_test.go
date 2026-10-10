package masking_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inouey1008/kusuri-api-poc/internal/masking"
)

func TestValues(t *testing.T) {
	t.Run(`秘密情報にあたる値を伏せる`, func(t *testing.T) {
		values := url.Values{"q": {"エゼチミブ"}, "token": {"abc123"}}

		assert.Equal(t, map[string][]string{
			"q":     {"エゼチミブ"},
			"token": {masking.MaskedValue},
		}, masking.Values(values))
	})

	t.Run(`複数の値を持つ場合もまとめて伏せる`, func(t *testing.T) {
		values := url.Values{"token": {"abc", "def"}}

		assert.Equal(t, map[string][]string{"token": {masking.MaskedValue}}, masking.Values(values))
	})
}

func TestJSON(t *testing.T) {
	t.Run(`オブジェクトのキーを見て伏せる`, func(t *testing.T) {
		input := map[string]any{"name": "エゼチミブ", "password": "hunter2"}

		assert.Equal(t, map[string]any{
			"name":     "エゼチミブ",
			"password": masking.MaskedValue,
		}, masking.JSON(input))
	})

	t.Run(`入れ子の中まで辿る`, func(t *testing.T) {
		input := map[string]any{
			"user": map[string]any{
				"id":            "u1",
				"refresh_token": "abc123",
			},
		}

		assert.Equal(t, map[string]any{
			"user": map[string]any{
				"id":            "u1",
				"refresh_token": masking.MaskedValue,
			},
		}, masking.JSON(input))
	})

	t.Run(`配列の要素も辿る`, func(t *testing.T) {
		input := map[string]any{
			"items": []any{
				map[string]any{"apiKey": "k1"},
				map[string]any{"name": "エゼチミブ"},
			},
		}

		assert.Equal(t, map[string]any{
			"items": []any{
				map[string]any{"apiKey": masking.MaskedValue},
				map[string]any{"name": "エゼチミブ"},
			},
		}, masking.JSON(input))
	})

	t.Run(`オブジェクト以外はそのまま返す`, func(t *testing.T) {
		assert.Equal(t, "エゼチミブ", masking.JSON("エゼチミブ"))
		assert.InEpsilon(t, 1.5, masking.JSON(1.5), 0.0001)
		assert.Nil(t, masking.JSON(nil))
	})
}

func TestIsSensitive(t *testing.T) {
	testCases := []struct {
		key       string
		sensitive bool
	}{
		{key: "token", sensitive: true},
		{key: "access_token", sensitive: true},
		{key: "API_KEY", sensitive: true},
		{key: "Authorization", sensitive: true},
		{key: "q", sensitive: false},
		{key: "keyword", sensitive: false},
		{key: "name", sensitive: false},

		// 旧 sentryx.ScrubEvent は Authorization/Cookie/Set-Cookie の完全一致 3 つだけを
		// マスクしていた。部分一致 (IsSensitive) 化した現在は "cookie" を含む名前なら
		// Set-Cookie も拾える (後退していないことの確認)。
		{key: "set-cookie", sensitive: true},
		{key: "Set-Cookie", sensitive: true},

		// ありがちな API キーヘッダ "X-Api-Key" はハイフン区切りだが、
		// IsSensitive は比較前に "-" と "_" を除いて正規化するため、
		// "apikey" に一致してマスク対象になる。
		{key: "X-Api-Key", sensitive: true},
		{key: "x-api-key", sensitive: true},

		// 区切りがアンダースコアやなしでも同様に一致する
		{key: "api_key", sensitive: true},
		{key: "ApiKey", sensitive: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.key, func(t *testing.T) {
			assert.Equal(t, testCase.sensitive, masking.IsSensitive(testCase.key))
		})
	}
}
