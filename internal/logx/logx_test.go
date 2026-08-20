package logx_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inouey1008/kusuri-api-poc/internal/logx"
)

// Init を適用した状態で log を実行し、標準出力に書かれた内容を返す
func captureOutput(t *testing.T, log func()) []byte {
	t.Helper()

	// writer に書いたものが reader から読めるようにする
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	originalStdout := os.Stdout
	originalLogger := slog.Default()

	// os.Stdout も slog のデフォルトもプロセス全体で共有される
	// 戻さないと後続のテストの出力までこのパイプに流れ込むため、
	// テスト完了後にクリーンアップするようにする
	t.Cleanup(func() {
		os.Stdout = originalStdout
		slog.SetDefault(originalLogger)
	})

	// ログは標準出力に書かれ、通常は読み返せない
	// os.Stdout をパイプに差し替えて検証可能にする
	os.Stdout = writer

	// !! Init は差し替えより後に呼ぶ必要あり
	logx.Init()

	log()

	// 閉じないと次の io.ReadAll が EOF を待ち続ける
	require.NoError(t, writer.Close())

	// 横取りした出力を全部読み出す
	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	return output
}

func TestInit(t *testing.T) {
	output := captureOutput(t, func() {
		slog.Debug("dropped")
		slog.Info("hello", slog.String("key", "value"))
	})

	t.Run(`Info より下のレベルは出力しない`, func(t *testing.T) {
		// slog は 1 レコードを 1 行で書く。Debug が出ていれば 2 行になる
		assert.Equal(t, 1, bytes.Count(output, []byte("\n")))
	})

	// JSON として解釈できなければ、ここで失敗する
	var entry map[string]any
	require.NoError(t, json.Unmarshal(output, &entry), "JSON として解釈できる形式で出力する")

	t.Run(`メッセージとレベルを記録する`, func(t *testing.T) {
		assert.Equal(t, "hello", entry["msg"])
		assert.Equal(t, "INFO", entry["level"])
		assert.NotEmpty(t, entry["time"])
	})

	t.Run(`属性はキーと値の組で出力する`, func(t *testing.T) {
		assert.Equal(t, "value", entry["key"])
	})
}
