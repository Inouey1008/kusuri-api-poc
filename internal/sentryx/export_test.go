package sentryx

// テストから非公開の関数を呼ぶための公開名
// _test.go なので本番のビルドには含まれない。
var MaskEvent = maskEvent
