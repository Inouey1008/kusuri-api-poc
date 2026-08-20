package errorx

import (
	"net/http"

	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

// エラーレスポンス
// - Status : HTTPErrorHandler が応答コードに使う。ボディには含めない
// - Code   : 言語非依存の識別子。Message が変わってもクライアントが判定できるようにする
// - Message: 人間向け文言。error インターフェースの戻り値も兼ねる
type Errorx struct {
	Status  int                     `json:"-"`
	Code    string                  `json:"code"`
	Message string                  `json:"error"`
	Details []validation.FieldError `json:"details,omitempty"`
}

// ハンドラから return できるよう error を満たす。
// 応答への変換は HTTPErrorHandler が一箇所で行う。
func (e Errorx) Error() string {
	return e.Message
}

// Details を上書きした Errorx のコピーを返す。Errorx 自体は変更しない。
func (e Errorx) WithDetails(details []validation.FieldError) Errorx {
	e.Details = details
	return e
}

var (
	// 500
	Internal = Errorx{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: "サーバー内部でエラーが発生しました"}
	// 400 (Details は WithDetails で追加すること)
	Validation = Errorx{Status: http.StatusBadRequest, Code: "VALIDATION_FAILED", Message: "入力内容に誤りがあります"}
)
