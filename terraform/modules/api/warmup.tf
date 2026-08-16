# コールドスタートを避けるため実行環境を温存する。効くのは同時実行 1 枠分のみ
resource "aws_scheduler_schedule" "warmup" {
  count = var.warmup_enabled ? 1 : 0

  name                = "${var.shared_prefix}-api-warmup"
  state               = "ENABLED"
  schedule_expression = "rate(${var.warmup_interval_minutes} minutes)"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    # Function URL と同じ経路を温めるため $LATEST ではなくエイリアスを叩く
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

    # 次の実行で叩き直されるため再試行は不要
    retry_policy {
      maximum_retry_attempts = 0
    }
  }
}

resource "aws_iam_role" "scheduler" {
  count = var.warmup_enabled ? 1 : 0

  name = "${var.shared_prefix}-api-warmup-scheduler"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Action    = "sts:AssumeRole"
      Principal = { Service = "scheduler.amazonaws.com" }
    }]
  })
}

# Scheduler はリソースベースポリシーではなく実行ロールで呼び出す
resource "aws_iam_role_policy" "scheduler_invoke" {
  count = var.warmup_enabled ? 1 : 0

  name = "invoke-alias"
  role = aws_iam_role.scheduler[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "lambda:InvokeFunction"
      Resource = aws_lambda_alias.current.arn
    }]
  })
}
