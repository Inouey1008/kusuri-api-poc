package middleware

import (
	"net/http"
	"net/url"

	"github.com/inouey1008/kusuri-api-poc/internal/masking"
)

// ログに残すヘッダ。増やすときは秘密情報が入らないか確認する
var loggedHeaders = []string{
	"User-Agent",
	"Referer",
	"Content-Type",
	"X-Forwarded-For",
}

// 記録するヘッダだけを取り出す。
// 絞ったうえでマスクも通し、選定を誤っても漏れないようにする。
func pickHeaders(header http.Header) map[string][]string {
	picked := make(http.Header, len(loggedHeaders))

	for _, name := range loggedHeaders {
		if values := header.Values(name); len(values) > 0 {
			picked[http.CanonicalHeaderKey(name)] = values
		}
	}

	return masking.Values(url.Values(picked))
}
