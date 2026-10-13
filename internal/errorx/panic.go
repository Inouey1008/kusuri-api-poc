package errorx

import (
	"fmt"
	"runtime"
)

// Panic は recover した panic を error として運ぶ。
// Recovery で値とスタックを保持しておくことで、送信を errorHandler 側の 1 箇所にまとめられる。
// 応答には使わない。errorHandler はこれを判別せず、既定の Internal (汎用の 500) に倒す。
type Panic struct {
	Value any
	stack []uintptr
}

// NewPanic は呼び出し元のスタックを記録した Panic を作る。
// panic を recover した defer の中で呼ぶこと。そこはまだ panic した goroutine のスタック上にいる。
func NewPanic(value any) *Panic {
	pcs := make([]uintptr, 64)
	n := runtime.Callers(2, pcs) // runtime.Callers と NewPanic 自身を除く

	return &Panic{Value: value, stack: pcs[:n]}
}

func (p *Panic) Error() string {
	return fmt.Sprintf("panic: %v", p.Value)
}

// StackTrace は sentry-go がスタックを取り出すために reflection で呼ぶ。
// メソッド名と戻り値の型は sentry-go の都合で決まっているため、変えないこと。
func (p *Panic) StackTrace() []uintptr {
	return p.stack
}
