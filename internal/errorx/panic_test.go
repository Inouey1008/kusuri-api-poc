package errorx_test

import (
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/errorx"
)

func TestPanic_Error(t *testing.T) {
	t.Run(`"panic: " に panic の値を続けた文字列を返す`, func(t *testing.T) {
		assert.EqualError(t, errorx.NewPanic("boom"), "panic: boom")
	})
}

func TestPanic_StackTrace(t *testing.T) {
	t.Run(`スタックトレースが空でない`, func(t *testing.T) {
		assert.NotEmpty(t, errorx.NewPanic("boom").StackTrace())
	})
}

// このテストは sentry-go の内部挙動への依存を検知するために置いている。
//
// Panic はスタックを Sentry へ届けるために、StackTrace() []uintptr というメソッドを持つ。
// sentry-go はこのメソッド名を reflection で探してスタックを取り出す (公開された契約ではない)。
// sentry-go の更新でこの挙動が変わると、Sentry へ送るイベントからスタックが黙って消える。
// 公開関数 sentry.ExtractStacktrace を直接呼び、実際にスタックを読めることをここで固定する。
func TestPanic_SentryExtractStacktrace(t *testing.T) {
	t.Run(`sentry-go がスタックを読み取れ、呼び出し元のテスト関数がフレームに含まれる`, func(t *testing.T) {
		stacktrace := sentry.ExtractStacktrace(errorx.NewPanic("boom"))

		require.NotNil(t, stacktrace, "sentry-go が StackTrace メソッドを認識できなくなった可能性がある")
		require.NotEmpty(t, stacktrace.Frames)

		functions := make([]string, 0, len(stacktrace.Frames))
		for _, frame := range stacktrace.Frames {
			functions = append(functions, frame.Function)
		}

		assert.Contains(t, functions, "TestPanic_SentryExtractStacktrace.func1",
			"NewPanic を呼んだ位置のスタックが記録されている")
	})
}
