resource "aws_sns_topic" "alerts" {
  name = "${var.shared_prefix}-alerts"
}

# 確認メールを踏むまで有効にならない
resource "aws_sns_topic_subscription" "email" {
  for_each = toset(var.alert_emails)

  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = each.value
}

resource "aws_cloudwatch_metric_alarm" "lambda_throttle" {
  alarm_name          = "${var.shared_prefix}-lambda-throttle"
  alarm_description   = "Lambda が同時実行の上限で拒否した。攻撃か上限不足のどちらか"
  namespace           = "AWS/Lambda"
  metric_name         = "Throttles"
  dimensions          = { FunctionName = aws_lambda_function.api.function_name }
  statistic           = "Sum"
  period              = 300
  evaluation_periods  = 1
  comparison_operator = "GreaterThanThreshold"
  threshold           = 0
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "request_spike" {
  alarm_name          = "${var.shared_prefix}-request-spike"
  alarm_description   = "リクエストが急増した"
  namespace           = "AWS/ApiGateway"
  metric_name         = "Count"
  dimensions          = { ApiId = aws_apigatewayv2_api.api.id }
  statistic           = "Sum"
  period              = 300
  evaluation_periods  = 1
  comparison_operator = "GreaterThanThreshold"
  threshold           = var.request_spike_threshold
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.alerts.arn]
}
