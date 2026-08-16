package health

// GET /health のレスポンス
type healthResponse struct {
	Status string `json:"status"`
}
