package drug

// drugResponse は Drug の JSON 転送表現。
// ドメイン層に JSON タグを持ち込まないよう handler 側で変換する。
type drugResponse struct {
	YJCode string `json:"yjCode"`
	Name   string `json:"name"`
}

func (d Drug) toResponse() drugResponse {
	return drugResponse{YJCode: d.YJCode, Name: d.Name}
}

// searchResponse は GET /drugs のレスポンス形状。
type searchResponse struct {
	Total int            `json:"total"`
	Items []drugResponse `json:"items"`
}

// searchRequest は GET /drugs の入力パラメータ。
type searchRequest struct {
	Q string `json:"q" validate:"omitempty,max=100"`
}

// getRequest は GET /drugs/{yjCode} の入力パラメータ。
type getRequest struct {
	YJCode string `json:"yjCode" validate:"required,len=12,alphanum"`
}
