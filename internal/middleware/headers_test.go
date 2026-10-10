package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
