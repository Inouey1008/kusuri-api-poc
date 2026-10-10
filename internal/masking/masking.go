// Package masking は秘密情報を伏せるための共通の判定基準を提供する。
// ログ出力と Sentry 送信の双方から参照する。基準が分かれると、
// 一方だけ伏せ忘れてもう一方から漏れるため、ここに一本化している。
package masking

import (
	"net/url"
	"strings"
)

// MaskedValue は伏せた値の表現。
const MaskedValue = "***"

// 値を伏せるパラメータ名 (区切りを除いた部分一致で判定)
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
}

// キー名の区切りは場所ごとに流儀が違う (X-Api-Key / api_key / apiKey)。
// 区切りを落としてから突き合わせ、どの流儀でも同じ判定になるようにする。
var separatorReplacer = strings.NewReplacer("-", "", "_", "")

// IsSensitive はキー名が秘密情報にあたるかを判定する。
func IsSensitive(key string) bool {
	normalized := separatorReplacer.Replace(strings.ToLower(key))

	for _, marker := range sensitiveMarkers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}

	return false
}

// Values は秘密情報にあたるキーの値を伏せる。クエリ文字列とヘッダの双方に使う。
func Values(values url.Values) map[string][]string {
	masked := make(map[string][]string, len(values))

	for key, list := range values {
		if IsSensitive(key) {
			masked[key] = []string{MaskedValue}
			continue
		}

		masked[key] = list
	}

	return masked
}

// JSON は JSON を辿って秘密情報にあたるキーの値を伏せる。
// 入れ子と配列にも入るため、深い階層のトークンも拾える。
func JSON(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		masked := make(map[string]any, len(typed))
		for key, element := range typed {
			if IsSensitive(key) {
				masked[key] = MaskedValue
				continue
			}

			masked[key] = JSON(element)
		}

		return masked

	case []any:
		masked := make([]any, len(typed))
		for i, element := range typed {
			masked[i] = JSON(element)
		}

		return masked

	default:
		return value
	}
}
