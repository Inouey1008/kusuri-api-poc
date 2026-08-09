package drug

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inouey1008/kusuri-api-poc/internal/validation"
)

// service は Handler が依存するユースケースの境界 (consumer-defined)。
// テスト時はスタブに差し替えられる。
type service interface {
	Search(ctx context.Context, q string) ([]Drug, error)
	GetByYJCode(ctx context.Context, yjCode string) (*Drug, error)
}

// Handler は医薬品エンドポイントのハンドラ群。
type Handler struct {
	svc service
}

// NewHandler は service を受け取り Handler を返す。
func NewHandler(svc service) *Handler {
	return &Handler{svc: svc}
}

// Register は自身のルートを mux に登録する。router 側はモジュールを知らなくてよい。
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /drugs", h.Search)
	mux.HandleFunc("GET /drugs/{yjCode}", h.GetByYJCode)
}

// Search は GET /drugs?q=... を処理し、件数と一覧を JSON で返す。
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	if errs := validation.Validate(searchRequest{Q: q}); errs != nil {
		writeValidationError(w, errs)
		return
	}

	items, err := h.svc.Search(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	responses := make([]drugResponse, len(items))
	for i, d := range items {
		responses[i] = toResponse(d)
	}
	writeJSON(w, http.StatusOK, searchResponse{
		Total: len(responses),
		Items: responses,
	})
}

// GetByYJCode は GET /drugs/{yjCode} を処理する。見つからなければ 404。
func (h *Handler) GetByYJCode(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("yjCode")

	if errs := validation.Validate(getRequest{YJCode: code}); errs != nil {
		writeValidationError(w, errs)
		return
	}

	d, err := h.svc.GetByYJCode(r.Context(), code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	if d == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, toResponse(*d))
}

// writeValidationError は 400 バリデーションエラーレスポンスを書き込む共通ヘルパ。
func writeValidationError(w http.ResponseWriter, details []validation.FieldError) {
	writeJSON(w, http.StatusBadRequest, map[string]any{
		"error":   "validation failed",
		"details": details,
	})
}

// writeJSON はステータスコードと値を JSON でレスポンスに書き込む共通ヘルパ。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
