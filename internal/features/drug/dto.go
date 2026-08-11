package drug

// レスポンス用の Drug
type drugResponse struct {
	YJCode string `json:"yjCode"`
	Name   string `json:"name"`
}

func (d Drug) toResponse() drugResponse {
	return drugResponse{YJCode: d.YJCode, Name: d.Name}
}

// GET /drugs のレスポンス
type searchResponse struct {
	Total int            `json:"total"`
	Items []drugResponse `json:"items"`
}

// GET /drugs の入力パラメータ
type searchRequest struct {
	Q string `json:"q" validate:"omitempty,max=100"`
}
