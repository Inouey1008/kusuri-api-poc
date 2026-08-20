package middleware

import (
	"net/http"
	"net/url"
	"strings"
)

const maskedValue = "***"

// 値を伏せるパラメータ名 (部分一致で判定)
var sensitiveMarkers = []string{
	"token",
	"password",
	"passwd",
	"secret",
	"credential",
	"authorization",
	"cookie",
	"signature",
	"apikey",
	"api_key",
}

func sanitizeQuery(values url.Values) map[string][]string {
	return sanitizeValues(values)
}

// ログに残すヘッダ。増やすときは秘密情報が入らないか確認する
var loggedHeaders = []string{
	"User-Agent",
	"Referer",
	"Content-Type",
	"X-Forwarded-For",
}

// 記録するヘッダだけを取り出す。
// 絞ったうえでサニタイズも通し、選定を誤っても漏れないようにする。
func pickHeaders(header http.Header) map[string][]string {
	picked := make(http.Header, len(loggedHeaders))

	for _, name := range loggedHeaders {
		if values := header.Values(name); len(values) > 0 {
			picked[http.CanonicalHeaderKey(name)] = values
		}
	}

	return sanitizeHeader(picked)
}

// Authorization の Bearer トークンや Cookie をここで伏せる
func sanitizeHeader(header http.Header) map[string][]string {
	return sanitizeValues(url.Values(header))
}

func sanitizeValues(values url.Values) map[string][]string {
	sanitized := make(map[string][]string, len(values))

	for key, list := range values {
		if isSensitive(key) {
			sanitized[key] = []string{maskedValue}
			continue
		}

		sanitized[key] = list
	}

	return sanitized
}

// JSON を辿って秘密情報にあたるキーの値を伏せる。
// 入れ子と配列にも入るため、深い階層のトークンも拾える。
func sanitizeJSON(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		sanitized := make(map[string]any, len(typed))
		for key, element := range typed {
			if isSensitive(key) {
				sanitized[key] = maskedValue
				continue
			}

			sanitized[key] = sanitizeJSON(element)
		}

		return sanitized

	case []any:
		sanitized := make([]any, len(typed))
		for i, element := range typed {
			sanitized[i] = sanitizeJSON(element)
		}

		return sanitized

	default:
		return value
	}
}

func isSensitive(key string) bool {
	lowered := strings.ToLower(key)

	for _, marker := range sensitiveMarkers {
		if strings.Contains(lowered, marker) {
			return true
		}
	}

	return false
}
