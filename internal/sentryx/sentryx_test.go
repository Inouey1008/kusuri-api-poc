package sentryx_test

import (
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/sentryx"
)

func TestScrubEvent(t *testing.T) {
	t.Run(`Request が nil でも panic せずそのまま返す`, func(t *testing.T) {
		event := &sentry.Event{}

		var got *sentry.Event
		require.NotPanics(t, func() {
			got = sentryx.ScrubEvent(event, nil)
		})

		assert.Same(t, event, got)
	})

	t.Run(`Authorization・Cookie・Set-Cookie ヘッダをマスクし、Cookies は空にする`, func(t *testing.T) {
		event := &sentry.Event{
			Request: &sentry.Request{
				Cookies: "session=abc123",
				Headers: map[string]string{
					"Authorization": "Bearer secret-token",
					"Cookie":        "session=abc123",
					"Set-Cookie":    "session=abc123; HttpOnly",
					"User-Agent":    "go-test",
				},
			},
		}

		got := sentryx.ScrubEvent(event, nil)

		assert.Equal(t, "", got.Request.Cookies)
		assert.Equal(t, "***", got.Request.Headers["Authorization"])
		assert.Equal(t, "***", got.Request.Headers["Cookie"])
		assert.Equal(t, "***", got.Request.Headers["Set-Cookie"])
		assert.Equal(t, "go-test", got.Request.Headers["User-Agent"], "秘匿対象以外のヘッダはそのまま残す")
	})

	t.Run(`ヘッダ名の大文字小文字が異なっていてもマスクする`, func(t *testing.T) {
		event := &sentry.Event{
			Request: &sentry.Request{
				Headers: map[string]string{
					"authorization": "Bearer secret-token",
					"COOKIE":        "session=abc123",
				},
			},
		}

		got := sentryx.ScrubEvent(event, nil)

		assert.Equal(t, "***", got.Request.Headers["authorization"])
		assert.Equal(t, "***", got.Request.Headers["COOKIE"])
	})
}
