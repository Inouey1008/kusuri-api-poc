# 同時実行数の制限では流量を抑えられない
resource "aws_apigatewayv2_api" "api" {
  name          = "${var.shared_prefix}-api"
  protocol_type = "HTTP"

  dynamic "cors_configuration" {
    for_each = length(var.cors_allow_origins) > 0 ? [1] : []

    content {
      allow_origins = var.cors_allow_origins
      allow_methods = ["GET"]
      allow_headers = ["Content-Type", "X-Firebase-AppCheck"]
      max_age       = 86400
    }
  }
}

resource "aws_apigatewayv2_integration" "api" {
  api_id           = aws_apigatewayv2_api.api.id
  integration_type = "AWS_PROXY"

  # アプリの echoadapter.NewV2 と揃える
  payload_format_version = "2.0"
  integration_uri        = aws_lambda_alias.current.invoke_arn

  # 先に切ると 504 になり、アプリのエラー処理を通らない
  timeout_milliseconds = var.timeout * 1000 + 1000
}

resource "aws_apigatewayv2_route" "default" {
  api_id    = aws_apigatewayv2_api.api.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.api.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.api.id
  name        = "$default"
  auto_deploy = true

  default_route_settings {
    throttling_rate_limit  = var.throttling_rate_limit
    throttling_burst_limit = var.throttling_burst_limit
  }

  # 絞られた要求は Lambda に届かず、アプリのログに残らない
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      request_id = "$context.requestId"
      method     = "$context.httpMethod"
      path       = "$context.path"
      status     = "$context.status"
      latency    = "$context.responseLatency"
      error      = "$context.error.message"
    })
  }
}

# API Gateway は実行ロールを使わない
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowInvokeFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  qualifier     = aws_lambda_alias.current.name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.api.execution_arn}/*/*"
}

resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.shared_prefix}-api"
  retention_in_days = var.log_retention_in_days
}
