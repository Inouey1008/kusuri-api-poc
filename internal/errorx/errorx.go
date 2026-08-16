package errorx

import (
	"net/http"

	"github.com/inouey1008/kusuri-api-poc/internal/httpx"
	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

// エラーレスポンス
// - Status : 書き込み時のみ使い、ボディには含めない
// - Code : 言語非依存の識別子。Message が変わってもクライアントが判定できるようにする
// - Message: 人間向け文言
type Errorx struct {
	Status  int                     `json:"-"`
	Code    string                  `json:"code"`
	Message string                  `json:"error"`
	Details []validation.FieldError `json:"details,omitempty"`
}

// Errorx を ResponseWriter に書き込む
func (e Errorx) Write(w http.ResponseWriter) {
	httpx.WriteJSON(w, e.Status, e)
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
