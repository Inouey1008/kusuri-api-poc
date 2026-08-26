resource "aws_scheduler_schedule" "warmup" {
  count = var.warmup_enabled ? 1 : 0

  name                = "${var.shared_prefix}-api-warmup"
  state               = "ENABLED"
  schedule_expression = "rate(${var.warmup_interval_minutes} minutes)"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    # リクエストは current エイリアスに対して飛ぶので、関数名で叩かない ($LATEST が呼ばれてしまう)
    arn      = aws_lambda_alias.current.arn
    role_arn = aws_iam_role.scheduler[0].arn

    input = jsonencode({
      version = "2.0"
      rawPath = "/health"
      requestContext = {
        http = {
          method = "GET"
          path   = "/health"
        }
      }
    })

    retry_policy {
      # 次の実行で叩き直されるため、リトライ不要
      maximum_retry_attempts = 0
    }
  }

  description = "Lambda のコールドスタートを避けるため、定期実行してウォームアップしておく"
}

resource "aws_iam_role" "scheduler" {
  count = var.warmup_enabled ? 1 : 0

  name = "${var.shared_prefix}-api-warmup-scheduler"
  # IAM の description は ASCII しか受け付けない
  description = "Assumed by EventBridge Scheduler to warm up the API"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "AllowSchedulerAssumeRole"
      Effect    = "Allow"
      Action    = "sts:AssumeRole"
      Principal = { Service = "scheduler.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy" "scheduler_invoke" {
  count = var.warmup_enabled ? 1 : 0

  name = "invoke-alias"
  role = aws_iam_role.scheduler[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid      = "AllowInvokeCurrentAlias"
      Effect   = "Allow"
      Action   = "lambda:InvokeFunction"
      Resource = aws_lambda_alias.current.arn
    }]
  })
}
