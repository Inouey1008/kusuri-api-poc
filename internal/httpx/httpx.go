package httpx

import (
	"encoding/json"
	"net/http"
)

// ステータスコードと値を JSON で ResponseWriter に書き込む
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
