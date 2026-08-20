package middleware

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeQuery(t *testing.T) {
	t.Run(`秘密情報にあたる値を伏せる`, func(t *testing.T) {
		values := url.Values{"q": {"エゼチミブ"}, "token": {"abc123"}}

		assert.Equal(t, map[string][]string{
			"q":     {"エゼチミブ"},
			"token": {maskedValue},
		}, sanitizeQuery(values))
	})

	t.Run(`複数の値を持つ場合もまとめて伏せる`, func(t *testing.T) {
		values := url.Values{"token": {"abc", "def"}}

		assert.Equal(t, map[string][]string{"token": {maskedValue}}, sanitizeQuery(values))
	})
}

func TestSanitizeHeader(t *testing.T) {
	t.Run(`Authorization と Cookie を伏せる`, func(t *testing.T) {
		header := http.Header{}
		header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9.payload.signature")
		header.Set("Cookie", "session=abc123")
		header.Set("User-Agent", "curl/8.7.1")

		sanitized := sanitizeHeader(header)

		assert.Equal(t, []string{maskedValue}, sanitized["Authorization"])
		assert.Equal(t, []string{maskedValue}, sanitized["Cookie"])
		assert.Equal(t, []string{"curl/8.7.1"}, sanitized["User-Agent"], "秘密でないヘッダは残す")
	})
}

func TestSanitizeJSON(t *testing.T) {
	t.Run(`オブジェクトのキーを見て伏せる`, func(t *testing.T) {
		input := map[string]any{"name": "エゼチミブ", "password": "hunter2"}

		assert.Equal(t, map[string]any{
			"name":     "エゼチミブ",
			"password": maskedValue,
		}, sanitizeJSON(input))
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
				"refresh_token": maskedValue,
			},
		}, sanitizeJSON(input))
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
				map[string]any{"apiKey": maskedValue},
				map[string]any{"name": "エゼチミブ"},
			},
		}, sanitizeJSON(input))
	})

	t.Run(`オブジェクト以外はそのまま返す`, func(t *testing.T) {
		assert.Equal(t, "エゼチミブ", sanitizeJSON("エゼチミブ"))
		assert.InEpsilon(t, 1.5, sanitizeJSON(1.5), 0.0001)
		assert.Nil(t, sanitizeJSON(nil))
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.key, func(t *testing.T) {
			assert.Equal(t, testCase.sensitive, isSensitive(testCase.key))
		})
	}
}

func TestPickHeaders(t *testing.T) {
	t.Run(`記録対象のヘッダだけを残す`, func(t *testing.T) {
		header := http.Header{}
		header.Set("User-Agent", "curl/8.7.1")
		header.Set("X-Forwarded-For", "203.0.113.1")
		header.Set("Accept", "*/*")
		header.Set("X-Amzn-Trace-Id", "Root=1-...")
		header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiJ9")

		picked := pickHeaders(header)

		assert.Equal(t, map[string][]string{
			"User-Agent":      {"curl/8.7.1"},
			"X-Forwarded-For": {"203.0.113.1"},
		}, picked, "対象外のヘッダは Authorization も含めて落とす")
	})

	t.Run(`大文字小文字を問わず拾う`, func(t *testing.T) {
		header := http.Header{}
		header.Set("user-agent", "curl/8.7.1")

		assert.Equal(t, map[string][]string{"User-Agent": {"curl/8.7.1"}}, pickHeaders(header))
	})
}
