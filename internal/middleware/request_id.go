package middleware

import (
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// RequestID は X-Request-Id を決めてレスポンスヘッダに入れる。
// ヘッダがあれば採用し、なければ生成する。処理は echo 標準に任せている。
// Logging と sentryx.Report はレスポンスヘッダからこの値を読む。
var RequestID = echomiddleware.RequestID()
